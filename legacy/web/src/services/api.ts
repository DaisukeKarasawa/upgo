import axios from 'axios'

const api = axios.create({
  baseURL: '/api/v1',
  headers: {
    'Content-Type': 'application/json',
  },
})

export default api

export const getPRs = async (params?: {
  page?: number
  limit?: number
  state?: string
  author?: string
  search?: string
}) => {
  const response = await api.get('/prs', { params })
  return response.data
}

export const getPR = async (id: number) => {
  const response = await api.get(`/prs/${id}`)
  return response.data
}

export const sync = async () => {
  const response = await api.post('/sync')
  return response.data
}

export const syncPR = async (id: number) => {
  const response = await api.post(`/prs/${id}/sync`)
  return response.data
}

export const getSyncStatus = async () => {
  const response = await api.get('/sync/status')
  return response.data
}

export const getDashboardUpdateStatus = async () => {
  const response = await api.get('/updates/dashboard')
  return response.data
}

export const getPRUpdateStatus = async (prId: number) => {
  const response = await api.get(`/updates/pr/${prId}`)
  return response.data
}
