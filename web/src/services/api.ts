import axios from 'axios'

const api = axios.create({
  baseURL: '/api',
  headers: {
    'Content-Type': 'application/json',
  },
})

export default api

export const getChanges = async (params?: {
  page?: number
  limit?: number
  status?: string
  branch?: string
}) => {
  const response = await api.get('/changes', { params })
  return response.data
}

export const getChange = async (id: number) => {
  const response = await api.get(`/changes/${id}`)
  return response.data
}

export const getBranches = async () => {
  const response = await api.get('/branches')
  return response.data
}

export const getStatuses = async () => {
  const response = await api.get('/statuses')
  return response.data
}

export const sync = async () => {
  const response = await api.post('/sync')
  return response.data
}

export const syncChange = async (changeNumber: number) => {
  const response = await api.post(`/sync/change/${changeNumber}`)
  return response.data
}

export const checkUpdates = async () => {
  const response = await api.get('/sync/check')
  return response.data
}
