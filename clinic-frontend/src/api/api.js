import axios from 'axios'

const api = axios.create({
    baseURL: 'http://localhost:8080/api',
    headers: { 'Content-Type': 'application/json' }
})

export const getPatients  = ()        => api.get('/patients')
export const createPatient = (data)   => api.post('/patients', data)
export const getDoctors   = ()        => api.get('/doctors')
export const parseVisit   = (data)    => api.post('/parse', data)
export const getVisit     = (id)      => api.get(`/visits/${id}`)
export const getBill      = (id)      => api.get(`/visits/${id}/bill`)
export const markPaid     = (id)      => api.post(`/visits/${id}/bill/pay`)
export const downloadPDF  = (id)      => window.open(`http://localhost:8080/api/visits/${id}/pdf`, '_blank')
 
export default api