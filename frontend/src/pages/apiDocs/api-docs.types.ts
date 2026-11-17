export type ApiField = readonly [name: string, type: string, required: boolean, description: string]
export type ApiStage = 'core' | 'extension'
export type ApiAccess =
  'public' | 'session' | 'student' | 'teacher' | 'staff' | 'admin' | 'member' | 'refresh' | 'ticket'
export interface ApiSchema {
  name: string
  description: string
  fields: ApiField[]
  example: unknown
}
export interface ApiEndpoint {
  method: 'GET' | 'POST' | 'PUT' | 'PATCH' | 'DELETE'
  path: string
  title: string
  access: ApiAccess
  stage: ApiStage
  query?: ApiField[]
  body?: string
  headers?: ApiField[]
  response: {
    status: number
    schema?: string
    array?: boolean
    page?: boolean
    nullable?: boolean
    upgrade?: boolean
    binary?: string
    description?: string
  }
  errors?: string[]
  notes: string[]
}
export interface ApiGroup {
  id: string
  title: string
  description: string
  source: string
  endpoints: ApiEndpoint[]
}
