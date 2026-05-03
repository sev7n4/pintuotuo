package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pintuotuo/backend/config"
)

func GetRouteDecisionLogs(c *gin.Context) {
	if !ensureAdmin(c) {
		return
	}

	db := config.GetDB()
	if db == nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "Database connection error",
		})
		return
	}

	// 解析查询参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	merchantID, _ := strconv.Atoi(c.DefaultQuery("merchant_id", "0"))
	apiKeyID, _ := strconv.Atoi(c.DefaultQuery("api_key_id", "0"))
	startTime := c.Query("start_time")
	endTime := c.Query("end_time")
	strategy := c.Query("strategy")
	decisionResult := c.Query("decision_result")
	requestID := c.Query("request_id")

	// 构建查询
	query := `
		SELECT id, request_id, merchant_id, api_key_id, strategy_layer_goal, 
		       strategy_layer_input, strategy_layer_output,
		       decision_layer_candidates, decision_layer_output,
		       execution_layer_result, decision_duration_ms, decision_result, error_message,
		       created_at
		FROM routing_decision_logs
		WHERE 1=1
	`
	var args []interface{}
	argPos := 1

	if requestID != "" {
		query += ` AND request_id LIKE $` + strconv.Itoa(argPos)
		args = append(args, "%"+requestID+"%")
		argPos++
	}

	if merchantID > 0 {
		query += ` AND merchant_id = $` + strconv.Itoa(argPos)
		args = append(args, merchantID)
		argPos++
	}

	if apiKeyID > 0 {
		query += ` AND api_key_id = $` + strconv.Itoa(argPos)
		args = append(args, apiKeyID)
		argPos++
	}

	if strategy != "" {
		query += ` AND strategy_layer_goal = $` + strconv.Itoa(argPos)
		args = append(args, strategy)
		argPos++
	}

	if decisionResult != "" {
		query += ` AND decision_result = $` + strconv.Itoa(argPos)
		args = append(args, decisionResult)
		argPos++
	}

	if startTime != "" {
		query += ` AND created_at >= $` + strconv.Itoa(argPos)
		args = append(args, startTime)
		argPos++
	}

	if endTime != "" {
		query += ` AND created_at <= $` + strconv.Itoa(argPos)
		args = append(args, endTime)
		argPos++
	}

	// 添加排序和分页
	query += ` ORDER BY created_at DESC LIMIT $` + strconv.Itoa(argPos) + ` OFFSET $` + strconv.Itoa(argPos+1)
	args = append(args, pageSize, (page-1)*pageSize)

	// 执行查询
	rows, err := db.QueryContext(c.Request.Context(), query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "Failed to query route decision logs",
		})
		return
	}
	defer rows.Close()

	// 解析结果
	var logs []map[string]interface{}
	for rows.Next() {
		var log map[string]interface{} = make(map[string]interface{})
		var (
			id                     int
			requestID              string
			merchantID             *int
			apiKeyID               *int
			strategyLayerGoal      string
			strategyLayerInput     []byte
			strategyLayerOutput    []byte
			decisionLayerCandidates []byte
			decisionLayerOutput    []byte
			executionLayerResult   []byte
			decisionDurationMs     int
			decisionResult         string
			errorMessage           *string
			createdAt              time.Time
		)

		scanErr := rows.Scan(
			&id,
			&requestID,
			&merchantID,
			&apiKeyID,
			&strategyLayerGoal,
			&strategyLayerInput,
			&strategyLayerOutput,
			&decisionLayerCandidates,
			&decisionLayerOutput,
			&executionLayerResult,
			&decisionDurationMs,
			&decisionResult,
			&errorMessage,
			&createdAt,
		)
		if scanErr != nil {
			continue
		}

		// 解析JSON字段
		var strategyInput map[string]interface{}
		var strategyOutput map[string]interface{}
		var decisionCandidates []map[string]interface{}
		var decisionOutput map[string]interface{}
		var executionResult map[string]interface{}

		if len(strategyLayerInput) > 0 {
			json.Unmarshal(strategyLayerInput, &strategyInput)
		}

		if len(strategyLayerOutput) > 0 {
			json.Unmarshal(strategyLayerOutput, &strategyOutput)
		}

		if len(decisionLayerCandidates) > 0 {
			json.Unmarshal(decisionLayerCandidates, &decisionCandidates)
		}

		if len(decisionLayerOutput) > 0 {
			json.Unmarshal(decisionLayerOutput, &decisionOutput)
		}

		if len(executionLayerResult) > 0 {
			json.Unmarshal(executionLayerResult, &executionResult)
		}

		log["id"] = id
		log["request_id"] = requestID
		if merchantID != nil {
			log["merchant_id"] = *merchantID
		}
		if apiKeyID != nil {
			log["api_key_id"] = *apiKeyID
		}
		log["strategy_layer_goal"] = strategyLayerGoal
		log["strategy_layer_input"] = strategyInput
		log["strategy_layer_output"] = strategyOutput
		log["decision_layer_candidates"] = decisionCandidates
		log["decision_layer_output"] = decisionOutput
		log["execution_layer_result"] = executionResult
		log["decision_duration_ms"] = decisionDurationMs
		log["decision_result"] = decisionResult
		if errorMessage != nil {
			log["error_message"] = *errorMessage
		}
		log["created_at"] = createdAt

		logs = append(logs, log)
	}

	// 获取总数
	countQuery := `
		SELECT COUNT(*)
		FROM routing_decision_logs
		WHERE 1=1
	`
	countArgs := args[:len(args)-2] // 移除分页参数
	var total int
	err = db.QueryRowContext(c.Request.Context(), countQuery, countArgs...).Scan(&total)
	if err != nil {
		total = 0
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data": gin.H{
			"logs":  logs,
			"total": total,
			"page":  page,
			"size":  pageSize,
		},
	})
}

func GetRouteDecisionLog(c *gin.Context) {
	if !ensureAdmin(c) {
		return
	}

	logIDStr := c.Param("id")
	logID, err := strconv.Atoi(logIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "Invalid log ID",
		})
		return
	}

	db := config.GetDB()
	if db == nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "Database connection error",
		})
		return
	}

	query := `
		SELECT id, request_id, merchant_id, api_key_id, strategy_layer_goal, 
		       strategy_layer_input, strategy_layer_output,
		       decision_layer_candidates, decision_layer_output,
		       execution_layer_result, decision_duration_ms, decision_result, error_message,
		       created_at
		FROM routing_decision_logs
		WHERE id = $1
	`

	var log map[string]interface{} = make(map[string]interface{})
	var (
		id                     int
		requestID              string
		merchantID             *int
		apiKeyID               *int
		strategyLayerGoal      string
		strategyLayerInput     []byte
		strategyLayerOutput    []byte
		decisionLayerCandidates []byte
		decisionLayerOutput    []byte
		executionLayerResult   []byte
		decisionDurationMs     int
		decisionResult         string
		errorMessage           *string
		createdAt              time.Time
	)

	err = db.QueryRowContext(c.Request.Context(), query, logID).Scan(
		&id,
		&requestID,
		&merchantID,
		&apiKeyID,
		&strategyLayerGoal,
		&strategyLayerInput,
		&strategyLayerOutput,
		&decisionLayerCandidates,
		&decisionLayerOutput,
		&executionLayerResult,
		&decisionDurationMs,
		&decisionResult,
		&errorMessage,
		&createdAt,
	)

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    404,
			"message": "Log not found",
		})
		return
	}

	// 解析JSON字段
	var strategyInput map[string]interface{}
	var strategyOutput map[string]interface{}
	var decisionCandidates []map[string]interface{}
	var decisionOutput map[string]interface{}
	var executionResult map[string]interface{}

	if len(strategyLayerInput) > 0 {
		json.Unmarshal(strategyLayerInput, &strategyInput)
	}

	if len(strategyLayerOutput) > 0 {
		json.Unmarshal(strategyLayerOutput, &strategyOutput)
	}

	if len(decisionLayerCandidates) > 0 {
		json.Unmarshal(decisionLayerCandidates, &decisionCandidates)
	}

	if len(decisionLayerOutput) > 0 {
		json.Unmarshal(decisionLayerOutput, &decisionOutput)
	}

	if len(executionLayerResult) > 0 {
		json.Unmarshal(executionLayerResult, &executionResult)
	}

	log["id"] = id
	log["request_id"] = requestID
	if merchantID != nil {
		log["merchant_id"] = *merchantID
	}
	if apiKeyID != nil {
		log["api_key_id"] = *apiKeyID
	}
	log["strategy_layer_goal"] = strategyLayerGoal
	log["strategy_layer_input"] = strategyInput
	log["strategy_layer_output"] = strategyOutput
	log["decision_layer_candidates"] = decisionCandidates
	log["decision_layer_output"] = decisionOutput
	log["execution_layer_result"] = executionResult
	log["decision_duration_ms"] = decisionDurationMs
	log["decision_result"] = decisionResult
	if errorMessage != nil {
		log["error_message"] = *errorMessage
	}
	log["created_at"] = createdAt

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    log,
	})
}
