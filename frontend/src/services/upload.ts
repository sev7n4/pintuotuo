import api from './api';
import { APIResponse } from '@/types';

export interface UploadResponse {
  url: string;
  filename: string;
  size: number;
}

export const uploadService = {
  uploadFile: async (file: File, type: 'logo' | 'license' | 'idcard' | 'misc' = 'misc'): Promise<string> => {
    const formData = new FormData();
    formData.append('file', file);

    const response = await api.post<APIResponse<UploadResponse>>(`/upload?type=${type}`, formData, {
      headers: {
        'Content-Type': 'multipart/form-data',
      },
    });

    if (!response.data.data) {
      throw new Error('Upload failed');
    }

    return response.data.data.url;
  },
};
