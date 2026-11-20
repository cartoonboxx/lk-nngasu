import { useAuthStore } from '@/stores/auth'
import { ApiError } from './client'
import type { ServerMessage } from './cabinet.types'

export type ConnectionState = 'connecting' | 'connected' | 'disconnected'
export type ChatEvent = { type: 'message.created'; data: { message: ServerMessage } }

type Frame =
  | ChatEvent
  | {
      type: 'session.ready' | 'ack' | 'error'
      request_id?: string
      data?: unknown
      error?: { code: string; details?: { fields?: Record<string, string[]> } }
    }

export class ChatConnection {
  private socket: WebSocket | null = null
  private stopped = false
  private connecting = false
  private ready = false
  private retry = 0
  private timer = 0
  private handshakeTimer = 0
  private generation = 0
  private refreshNeeded = false
  private pending = new Map<
    string,
    { resolve: (value: unknown) => void; reject: (error: Error) => void; timer: number }
  >()
  private userId = useAuthStore().user?.id
  private resume = () => {
    if (!this.stopped && !this.socket && document.visibilityState === 'visible') {
      window.clearTimeout(this.timer)
      void this.start()
    }
  }

  constructor(
    private onState: (state: ConnectionState) => void,
    private onEvent: (event: ChatEvent) => void,
    private onReady: () => void,
  ) {
    window.addEventListener('online', this.resume)
    document.addEventListener('visibilitychange', this.resume)
  }

  async start() {
    if (this.stopped || this.connecting || this.socket) return
    const generation = this.generation
    this.connecting = true
    this.onState('connecting')
    try {
      const auth = useAuthStore()
      if (this.refreshNeeded) {
        this.refreshNeeded = false
        await auth.refresh()
      }
      if (this.stopped || generation !== this.generation || auth.user?.id !== this.userId) return
      const ticket = await auth.request<{ ticket: string; endpoint: string; protocol: string }>(
        '/realtime/tickets',
        { method: 'POST', notifyOnError: false },
      )
      if (this.stopped || generation !== this.generation || auth.user?.id !== this.userId) return
      const url = new URL(ticket.endpoint, window.location.origin)
      if (
        url.origin !== window.location.origin ||
        url.pathname !== '/api/v1/realtime' ||
        ticket.protocol !== 'lk-chat.v1'
      )
        throw new Error('Некорректный адрес соединения.')
      url.protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
      url.searchParams.set('ticket', ticket.ticket)
      const socket = new WebSocket(url, ticket.protocol)
      this.socket = socket
      this.handshakeTimer = window.setTimeout(() => socket.close(4000, 'ready_timeout'), 12000)
      socket.onmessage = (event: MessageEvent<string>) => {
        if (this.stopped || this.socket !== socket) return
        let frame: Frame
        try {
          frame = JSON.parse(event.data) as Frame
          if (!frame || typeof frame !== 'object' || typeof frame.type !== 'string')
            throw new Error('Invalid frame')
        } catch {
          socket.close(4000, 'invalid_frame')
          return
        }
        if (frame.type === 'session.ready') {
          window.clearTimeout(this.handshakeTimer)
          this.retry = 0
          this.ready = true
          this.onState('connected')
          this.onReady()
        } else if (frame.type === 'ack' || frame.type === 'error') {
          const pending = frame.request_id ? this.pending.get(frame.request_id) : undefined
          if (!pending || !frame.request_id) return
          window.clearTimeout(pending.timer)
          this.pending.delete(frame.request_id)
          if (frame.type === 'ack') pending.resolve(frame.data)
          else
            pending.reject(
              new ApiError(
                400,
                frame.error?.code ?? 'invalid_response',
                frame.error?.details?.fields,
              ),
            )
        } else if (this.ready && frame.type === 'message.created') {
          this.onEvent(frame)
        }
      }
      socket.onclose = (event) => {
        if (this.stopped || this.socket !== socket) return
        window.clearTimeout(this.handshakeTimer)
        this.socket = null
        this.ready = false
        this.onState('disconnected')
        for (const pending of this.pending.values()) {
          window.clearTimeout(pending.timer)
          pending.reject(new Error('Связь прервана. Проверьте историю перед повторной отправкой.'))
        }
        this.pending.clear()
        this.refreshNeeded = event.code === 4401
        if (event.code !== 4403) this.scheduleReconnect()
      }
      socket.onerror = () => {
        // Браузер сообщает подробности через onclose; там же выполняется переподключение.
      }
    } catch {
      if (!this.stopped && generation === this.generation) {
        this.onState('disconnected')
        if (useAuthStore().user?.id === this.userId) this.scheduleReconnect()
      }
    } finally {
      if (generation === this.generation) this.connecting = false
    }
  }

  private scheduleReconnect() {
    window.clearTimeout(this.timer)
    if (this.stopped) return
    const delay = Math.min(30000, 1000 * 2 ** Math.min(this.retry++, 5)) + Math.random() * 500
    this.timer = window.setTimeout(() => void this.start(), delay)
  }

  command<T>(type: string, data: Record<string, unknown>): Promise<T> {
    const socket = this.socket
    if (!this.ready || socket?.readyState !== WebSocket.OPEN)
      return Promise.reject(new Error('Нет связи с чатом. Дождитесь подключения.'))
    const requestId = crypto.randomUUID()
    return new Promise<T>((resolve, reject) => {
      const timer = window.setTimeout(() => {
        this.pending.delete(requestId)
        reject(new Error('Подтверждение не получено. Обновите историю перед повторной отправкой.'))
      }, 15000)
      this.pending.set(requestId, { resolve: (value) => resolve(value as T), reject, timer })
      try {
        socket.send(JSON.stringify({ type, request_id: requestId, data }))
      } catch {
        window.clearTimeout(timer)
        this.pending.delete(requestId)
        reject(new Error('Не удалось отправить команду. Проверьте соединение.'))
      }
    })
  }

  stop() {
    this.stopped = true
    this.generation++
    window.clearTimeout(this.timer)
    window.clearTimeout(this.handshakeTimer)
    window.removeEventListener('online', this.resume)
    document.removeEventListener('visibilitychange', this.resume)
    this.socket?.close(1000, 'session_ended')
    this.socket = null
    this.ready = false
    for (const pending of this.pending.values()) {
      window.clearTimeout(pending.timer)
      pending.reject(new Error('Сессия завершена.'))
    }
    this.pending.clear()
  }
}
