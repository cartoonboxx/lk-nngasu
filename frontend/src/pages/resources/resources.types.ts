import type { Component } from 'vue'

export type Service = {
  label: string
  description: string
  domain: string
  url: string
  linkText: string
  icon: string | Component
}

export type Contact = {
  label: string
  description: string
  url: string
  linkText: string
  icon: Component
}
