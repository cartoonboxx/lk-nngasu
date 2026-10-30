<template>
  <div
    :class="[style.mediaButton, readonly ? style.readonly : null, disabled ? style.disabled : null]"
  >
    <component
      v-press
      :is="readonly ? 'div' : 'button'"
      :class="style.main"
      :type="readonly ? undefined : type"
      :disabled="!readonly && disabled ? true : undefined"
      @click="download"
    >
      <span :class="style.leftPart">
        <span
          v-if="$slots.icon"
          :class="style.icon"
          :style="{ '--file-icon-color': fileType.color }"
        >
          <component :is="fileType.icon" :size="28" />
        </span>
        <span :class="style.content">
          <slot>
            <span :class="style.title">{{ title }}</span>
            <span v-if="meta" :class="style.meta">{{ meta }}</span>
          </slot>
        </span>
      </span>
      <PhDownloadSimple v-if="!readonly && !$slots.actions" :class="style.download" :size="20" />
    </component>
    <div v-if="$slots.actions" :class="style.actions" :inert="disabled || undefined">
      <slot name="actions" />
    </div>
  </div>
</template>

<script lang="ts">
import { pressFeedback } from '@/directives/pressFeedback'
import { defineComponent, markRaw, type PropType } from 'vue'
import { PhDownloadSimple, PhFile } from '@phosphor-icons/vue'
import style from './MediaButton.module.scss'
import { fileTypes, fileTypeColors } from '@/components/UI/buttons/MediaButton/files.types.ts'

export default defineComponent({
  directives: { press: pressFeedback },
  name: 'MediaButton',
  components: { PhDownloadSimple },
  props: {
    src: {
      type: [String, Blob] as PropType<string | Blob>,
      default: undefined,
    },
    filename: {
      type: String,
      default: '',
    },
    title: {
      type: String,
      default: '',
    },
    meta: {
      type: String,
      default: '',
    },
    readonly: {
      type: Boolean,
      default: false,
    },
    disabled: {
      type: Boolean,
      default: false,
    },
    type: {
      type: String as PropType<ButtonType>,
      default: 'button',
    },
  },
  emits: { click: (event: MouseEvent) => event instanceof MouseEvent },
  data() {
    return { style }
  },
  computed: {
    fileType() {
      const names = [
        this.filename,
        this.src instanceof File ? this.src.name : '',
        this.title,
        this.meta,
      ]
      const name = names.find((value) => /\.[a-z0-9]+$/i.test(value.trim())) ?? ''
      const extension = name.trim().split('.').at(-1)?.toLowerCase() as keyof typeof fileTypes
      return {
        icon: fileTypes[extension] ?? markRaw(PhFile),
        color: fileTypeColors[extension] ?? 'var(--color-text-secondary)',
      }
    },
  },
  methods: {
    download(event: MouseEvent) {
      if (this.readonly || this.disabled) return
      this.$emit('click', event)
      if (!this.src || event.defaultPrevented) return

      const source = this.src
      const url = source instanceof Blob ? URL.createObjectURL(source) : source
      const link = document.createElement('a')
      link.href = url

      link.download = this.filename || (source instanceof File ? source.name : this.title)
      try {
        document.body.append(link)
        link.click()
      } finally {
        link.remove()
        if (source instanceof Blob) setTimeout(() => URL.revokeObjectURL(url), 1000)
      }
    },
  },
})
</script>
