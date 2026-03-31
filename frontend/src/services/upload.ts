import api from './api';
import { APIResponse } from '@/types';

export interface UploadResponse {
  url: string;
  filename: string;
  size: number;
}

export const uploadService = {
  uploadFile: async (
    file: File,
    type: 'logo' | 'license' | 'idcard' | 'misc' = 'misc'
  ): Promise<string> => {
    const formData = new FormData();
    formData.append('file', file);

    const response = await api.post<APIResponse<UploadResponse> | UploadResponse>(
      `/upload?type=${type}`,
      formData,
      {
        headers: {
          'Content-Type': 'multipart/form-data',
        },
      }
    );

    const data = response.data;
    if ('data' in data && data.data?.url) {
      return data.data.url;
    }
    if ('url' in data && data.url) {
      return data.url;
    }

    throw new Error('Upload failed: invalid response format');
  },
};
