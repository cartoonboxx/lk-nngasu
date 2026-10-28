import type { ButtonHTMLAttributes } from 'vue'

declare global {
  type UiSize = 'sm' | 'md' | 'lg'
  type UiState = 'default' | 'hover' | 'active' | 'disabled'

  type ButtonSize = UiSize
  type ButtonState = UiState
  type ButtonType = 'button' | 'submit' | 'reset'
  type ButtonVariant = 'primary' | 'secondary' | 'ghost' | 'danger'

  type IconButtonSize = UiSize
  type IconButtonState = UiState
  type IconButtonVariant = ButtonVariant

  type IconTabSize = UiSize
  type IconTabState = UiState
  type IconTabVariant = 'default' | 'selected'

  type CheckboxSize = UiSize
  type CheckboxState = UiState | 'focus'
  type CheckboxVariant = 'default' | 'error'

  type InputSize = UiSize
  type InputType = 'text' | 'email' | 'password' | 'search' | 'tel' | 'url'
  type InputVariant = 'default' | 'filled' | 'error' | 'disabled'

  type TooltipPosition = 'top' | 'bottom' | 'left' | 'right'
  type FloatingPopupMode = 'tabs' | 'select'
  type SelectSize = UiSize
  type SelectVariant = Exclude<InputVariant, 'error'>
  type NavigationItemVariant = 'bottom' | 'list'
  type TabTriggerAttrs = ButtonHTMLAttributes & Record<string, unknown>
  type ModalAnchor = HTMLElement | (() => HTMLElement | null) | null
  type InputElement = HTMLInputElement | HTMLTextAreaElement

  interface ModalContentTouch {
    id: number
    x: number
    y: number
    time: number
  }

  interface FloatingPopupItem {
    value: string
    label: string
    disabled?: boolean
  }
}

export {}
