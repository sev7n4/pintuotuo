package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"strings"

	"github.com/pintuotuo/backend/models"
	"github.com/pintuotuo/backend/services"
	"github.com/pintuotuo/backend/utils"
)

// 1×1 PNG（透明），用于 images/variations、images/edits 的 multipart。
var probePNG1x1, _ = base64.StdEncoding.DecodeString(
	"iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mP8/x8AAwMB/6X9n0kAAAAASUVORK5CYII=",
)

func authTokenForProbe(cfg *services.ExecutionProviderConfig, decrypted string) string {
	if cfg == nil {
		return decrypted
	}
	if cfg.GatewayMode == services.GatewayModeLitellm {
		if mk := strings.TrimSpace(os.Getenv("LITELLM_MASTER_KEY")); mk != "" {
			return mk
		}
	}
	return decrypted
}

func writeProbeRow(w *csv.Writer, ts string, key *models.MerchantAPIKey, apiFormat, routeMode, probe, httpCode, ok, note string) {
	_ = w.Write([]string{
		ts, itoa(key.ID), itoa(key.MerchantID), key.Provider, apiFormat, routeMode,
		probe, httpCode, ok, truncate(note, 600),
	})
	w.Flush()
}

func pcmWavMono16LE(sampleRate, durationMs int) []byte {
	numSamples := sampleRate * durationMs / 1000
	if numSamples < 1 {
		numSamples = 1
	}
	dataSize := numSamples * 2
	buf := make([]byte, 44+dataSize)
	copy(buf[0:4], []byte("RIFF"))
	lePutUint32(buf[4:8], uint32(36+dataSize))
	copy(buf[8:12], []byte("WAVEfmt "))
	lePutUint32(buf[12:16], 16)
	lePutUint16(buf[16:18], 1)
	lePutUint16(buf[18:20], 1)
	lePutUint32(buf[20:24], uint32(sampleRate))
	lePutUint32(buf[24:28], uint32(sampleRate*2))
	lePutUint16(buf[28:30], 2)
	lePutUint16(buf[30:32], 16)
	copy(buf[32:36], []byte("data"))
	lePutUint32(buf[36:40], uint32(dataSize))
	// PCM silence already zero
	return buf
}

func lePutUint32(b []byte, v uint32) {
	b[0] = byte(v)
	b[1] = byte(v >> 8)
	b[2] = byte(v >> 16)
	b[3] = byte(v >> 24)
}

func lePutUint16(b []byte, v uint16) {
	b[0] = byte(v)
	b[1] = byte(v >> 8)
}

func postJSONProbe(ctx context.Context, client *http.Client, url, bearer, body string) (code int, ok bool, note string) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, strings.NewReader(body))
	if err != nil {
		return 0, false, err.Error()
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+bearer)
	resp, err := client.Do(req)
	if err != nil {
		return 0, false, "http:" + err.Error()
	}
	defer func() { _ = resp.Body.Close() }()
	_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 8192))
	return resp.StatusCode, resp.StatusCode >= 200 && resp.StatusCode < 300, resp.Status
}

func postMultipartProbe(ctx context.Context, client *http.Client, url, bearer string, build func(mw *multipart.Writer) error) (code int, ok bool, note string) {
	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	if err := build(mw); err != nil {
		return 0, false, "multipart_build:" + err.Error()
	}
	if err := mw.Close(); err != nil {
		return 0, false, "multipart_close:" + err.Error()
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, &buf)
	if err != nil {
		return 0, false, err.Error()
	}
	req.Header.Set("Content-Type", mw.FormDataContentType())
	req.Header.Set("Authorization", "Bearer "+bearer)
	resp, err := client.Do(req)
	if err != nil {
		return 0, false, "http:" + err.Error()
	}
	defer func() { _ = resp.Body.Close() }()
	_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 8192))
	return resp.StatusCode, resp.StatusCode >= 200 && resp.StatusCode < 300, resp.Status
}

// runOpenAIFormatProbes emits one CSV row per OpenAI-shaped endpoint (与 services.EndpointType* 对齐)。
func runOpenAIFormatProbes(
	ctx context.Context,
	w *csv.Writer,
	ts string,
	client *http.Client,
	longClient *http.Client,
	key *models.MerchantAPIKey,
	apiFormat, routeMode string,
	mpCode, mpAPIBase, mpProviderRegion string,
	mpEndpoints, mpRouteStrategy []byte,
	p probeFlags,
) {
	cfg, err := buildExecCfg(mpCode, mpAPIBase, apiFormat, mpProviderRegion, mpRouteStrategy, mpEndpoints, key)
	if err != nil {
		writeProbeRow(w, ts, key, apiFormat, routeMode, "openai_probe_init", "0", "false", "build_cfg:"+err.Error())
		return
	}
	decrypted, derr := utils.Decrypt(key.APIKeyEncrypted)
	if derr != nil || decrypted == "" {
		writeProbeRow(w, ts, key, apiFormat, routeMode, "openai_probe_init", "0", "false", "decrypt_failed")
		return
	}
	tok := authTokenForProbe(cfg, decrypted)

	baseURL, _, err := services.ResolveMerchantAPIKeyUpstreamBase(ctx, key)
	if err != nil {
		writeProbeRow(w, ts, key, apiFormat, routeMode, "resolve_base", "0", "false", err.Error())
		return
	}
	chatURL := services.OpenAICompatChatCompletionsURL(baseURL)

	// --- embeddings ---
	if p.skipEmbeddings {
		writeProbeRow(w, ts, key, apiFormat, routeMode, "post_"+services.EndpointTypeEmbeddings, "0", "skipped", "skip_embeddings_flag")
	} else {
		u := services.ResolveEndpointByType(cfg, services.EndpointTypeEmbeddings)
		body := fmt.Sprintf(`{"model":%q,"input":"probe"}`, p.embeddingModel)
		code, ok, note := postJSONProbe(ctx, client, u, tok, body)
		writeProbeRow(w, ts, key, apiFormat, routeMode, "post_"+services.EndpointTypeEmbeddings, itoa(code), boolStr(ok), note)
	}

	// --- moderations（通常低成本）---
	{
		u := services.ResolveEndpointByType(cfg, services.EndpointTypeModerations)
		body := fmt.Sprintf(`{"model":%q,"input":"probe"}`, p.moderationModel)
		code, ok, note := postJSONProbe(ctx, client, u, tok, body)
		writeProbeRow(w, ts, key, apiFormat, routeMode, "post_"+services.EndpointTypeModerations, itoa(code), boolStr(ok), note)
	}

	// --- chat completions ---
	if !p.billable {
		writeProbeRow(w, ts, key, apiFormat, routeMode, "post_"+services.EndpointTypeChatCompletions, "0", "skipped", "set_-billable_to_enable_token_charging_probes")
	} else {
		body := fmt.Sprintf(`{"model":%q,"messages":[{"role":"user","content":"p"}],"max_tokens":1}`, p.chatModel)
		code, ok, note := postJSONProbe(ctx, longClient, chatURL, tok, body)
		writeProbeRow(w, ts, key, apiFormat, routeMode, "post_"+services.EndpointTypeChatCompletions, itoa(code), boolStr(ok), note)
	}

	// --- images generations ---
	if !p.billable {
		writeProbeRow(w, ts, key, apiFormat, routeMode, "post_"+services.EndpointTypeImagesGenerations, "0", "skipped", "billable_disabled")
	} else {
		u := services.ResolveEndpointByType(cfg, services.EndpointTypeImagesGenerations)
		body := fmt.Sprintf(`{"model":%q,"prompt":"solid gray","n":1,"size":"256x256"}`, p.imageModel)
		code, ok, note := postJSONProbe(ctx, longClient, u, tok, body)
		writeProbeRow(w, ts, key, apiFormat, routeMode, "post_"+services.EndpointTypeImagesGenerations, itoa(code), boolStr(ok), note)
	}

	// --- images variations (multipart) ---
	if !p.billable {
		writeProbeRow(w, ts, key, apiFormat, routeMode, "post_"+services.EndpointTypeImagesVariations, "0", "skipped", "billable_disabled")
	} else {
		u := services.ResolveEndpointByType(cfg, services.EndpointTypeImagesVariations)
		code, ok, note := postMultipartProbe(ctx, longClient, u, tok, func(mw *multipart.Writer) error {
			_ = mw.WriteField("n", "1")
			_ = mw.WriteField("size", "256x256")
			_ = mw.WriteField("model", "dall-e-2")
			fw, err := mw.CreateFormFile("image", "probe.png")
			if err != nil {
				return err
			}
			_, err = fw.Write(probePNG1x1)
			return err
		})
		writeProbeRow(w, ts, key, apiFormat, routeMode, "post_"+services.EndpointTypeImagesVariations, itoa(code), boolStr(ok), note)
	}

	// --- images edits ---
	if !p.billable {
		writeProbeRow(w, ts, key, apiFormat, routeMode, "post_"+services.EndpointTypeImagesEdits, "0", "skipped", "billable_disabled")
	} else {
		u := services.ResolveEndpointByType(cfg, services.EndpointTypeImagesEdits)
		code, ok, note := postMultipartProbe(ctx, longClient, u, tok, func(mw *multipart.Writer) error {
			_ = mw.WriteField("prompt", "make it slightly lighter")
			_ = mw.WriteField("model", "dall-e-2")
			_ = mw.WriteField("n", "1")
			_ = mw.WriteField("size", "256x256")
			fw, err := mw.CreateFormFile("image", "probe.png")
			if err != nil {
				return err
			}
			_, err = fw.Write(probePNG1x1)
			return err
		})
		writeProbeRow(w, ts, key, apiFormat, routeMode, "post_"+services.EndpointTypeImagesEdits, itoa(code), boolStr(ok), note)
	}

	wav := pcmWavMono16LE(16000, 200)

	// --- audio transcriptions ---
	if !p.billable {
		writeProbeRow(w, ts, key, apiFormat, routeMode, "post_"+services.EndpointTypeAudioTranscriptions, "0", "skipped", "billable_disabled")
	} else {
		u := services.ResolveEndpointByType(cfg, services.EndpointTypeAudioTranscriptions)
		code, ok, note := postMultipartProbe(ctx, longClient, u, tok, func(mw *multipart.Writer) error {
			_ = mw.WriteField("model", p.transcriptionModel)
			fw, err := mw.CreateFormFile("file", "probe.wav")
			if err != nil {
				return err
			}
			_, err = fw.Write(wav)
			return err
		})
		writeProbeRow(w, ts, key, apiFormat, routeMode, "post_"+services.EndpointTypeAudioTranscriptions, itoa(code), boolStr(ok), note)
	}

	// --- audio translations ---
	if !p.billable {
		writeProbeRow(w, ts, key, apiFormat, routeMode, "post_"+services.EndpointTypeAudioTranslations, "0", "skipped", "billable_disabled")
	} else {
		u := services.ResolveEndpointByType(cfg, services.EndpointTypeAudioTranslations)
		code, ok, note := postMultipartProbe(ctx, longClient, u, tok, func(mw *multipart.Writer) error {
			_ = mw.WriteField("model", p.transcriptionModel)
			_ = mw.WriteField("prompt", "translate to English")
			fw, err := mw.CreateFormFile("file", "probe.wav")
			if err != nil {
				return err
			}
			_, err = fw.Write(wav)
			return err
		})
		writeProbeRow(w, ts, key, apiFormat, routeMode, "post_"+services.EndpointTypeAudioTranslations, itoa(code), boolStr(ok), note)
	}

	// --- audio speech ---
	if !p.billable {
		writeProbeRow(w, ts, key, apiFormat, routeMode, "post_"+services.EndpointTypeAudioSpeech, "0", "skipped", "billable_disabled")
	} else {
		u := services.ResolveEndpointByType(cfg, services.EndpointTypeAudioSpeech)
		body := fmt.Sprintf(`{"model":%q,"input":"hi","voice":"alloy"}`, p.speechModel)
		code, ok, note := postJSONProbe(ctx, longClient, u, tok, body)
		writeProbeRow(w, ts, key, apiFormat, routeMode, "post_"+services.EndpointTypeAudioSpeech, itoa(code), boolStr(ok), note)
	}

	// --- responses ---
	{
		u := services.ResolveEndpointByType(cfg, services.EndpointTypeResponses)
		maxOut := 1
		bodyStruct := map[string]interface{}{
			"model":             p.chatModel,
			"input":             "ping",
			"max_output_tokens": maxOut,
		}
		b, _ := json.Marshal(bodyStruct)
		code, ok, note := postJSONProbe(ctx, longClient, u, tok, string(b))
		writeProbeRow(w, ts, key, apiFormat, routeMode, "post_"+services.EndpointTypeResponses, itoa(code), boolStr(ok), note)
	}
}

type probeFlags struct {
	skipEmbeddings   bool
	billable         bool
	embeddingModel   string
	moderationModel  string
	chatModel        string
	imageModel       string
	speechModel      string
	transcriptionModel string
}

func boolStr(v bool) string {
	if v {
		return "true"
	}
	return "false"
}
