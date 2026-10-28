import { type Component, type PropType } from 'vue'

export const baseTabProps = {
  triggerAttrs: {
    type: Object as PropType<TabTriggerAttrs>,
    default: () => ({}),
  },
  text: {
    type: String,
    default: '',
  },
  imageSrc: {
    type: String,
    default: undefined,
  },
  imageAlt: {
    type: String,
    default: '',
  },
  icon: {
    type: [Object, Function] as PropType<Component>,
    default: undefined,
  },
  iconSize: {
    type: Number,
    default: 22,
  },
  iconWeight: {
    type: String,
    default: 'regular',
  },
  active: {
    type: Boolean,
    default: false,
  },
  disabled: {
    type: Boolean,
    default: false,
  },
} as const
