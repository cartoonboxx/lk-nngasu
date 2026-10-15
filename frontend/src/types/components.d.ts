declare global {
  type ButtonSize = 'sm' | 'md' | 'lg'
  type ButtonState = 'default' | 'hover' | 'active' | 'disabled'
  type ButtonType = 'button' | 'submit' | 'reset'
  type ButtonVariant = 'primary' | 'secondary' | 'ghost' | 'danger'

  type IconButtonSize = 'sm' | 'md' | 'lg'
  type IconButtonState = 'default' | 'hover' | 'active' | 'disabled'
  type IconButtonVariant = 'primary' | 'secondary' | 'ghost' | 'danger'

  type IconTabSize = 'sm' | 'md' | 'lg'
  type IconTabState = 'default' | 'hover' | 'active' | 'disabled'
  type IconTabVariant = 'default' | 'selected'

  type CheckboxSize = 'sm' | 'md' | 'lg'
  type CheckboxState = 'default' | 'hover' | 'focus' | 'active' | 'disabled'
  type CheckboxVariant = 'default' | 'error'

  type InputSize = 'sm' | 'md' | 'lg'
  type InputType = 'text' | 'email' | 'password' | 'search' | 'tel' | 'url'
  type InputVariant = 'default' | 'filled' | 'error' | 'disabled'
}

export {}
