<template>
  <div :class="style.form">
    <label :class="style.field">
      {{ label }}
      <input type="file" accept="image/png,image/jpeg,image/webp" @change="readImage" />
    </label>
    <p :class="style.muted">PNG, JPEG или WebP, до 2 МБ. Необязательное изображение.</p>
    <p v-if="error" :class="style.error" role="alert">{{ error }}</p>
    <div v-if="modelValue" :class="style.actions">
      <img
        :src="modelValue"
        alt="Выбранное изображение"
        width="96"
        height="72"
        :class="style.imageThumbnail"
      />
      <Button variant="ghost" size="sm" @click="$emit('update:modelValue', '')">
        Убрать изображение
      </Button>
    </div>
  </div>
</template>
<script lang="ts">
import { defineComponent } from 'vue'
import Button from '@/components/UI/buttons/Button/Button.vue'
import style from '../Admin.module.scss'
export default defineComponent({
  components: { Button },
  props: {
    modelValue: {
      type: String,
      default: '',
    },
    label: {
      type: String,
      default: 'Обложка',
    },
  },
  emits: ['update:modelValue'],
  data() {
    return { style, error: '', readVersion: 0 }
  },
  beforeUnmount() {
    this.readVersion++
  },
  methods: {
    async readImage(event: Event) {
      const input = event.target as HTMLInputElement
      const file = input.files?.[0]
      if (!file) return
      const version = ++this.readVersion
      this.error = ''
      if (
        !['image/png', 'image/jpeg', 'image/webp'].includes(file.type) ||
        file.size > 2 * 1024 * 1024
      ) {
        this.error = 'Выберите PNG, JPEG или WebP размером до 2 МБ.'
        input.value = ''
        return
      }
      try {
        const result = await new Promise<string>((resolve, reject) => {
          const reader = new FileReader()
          reader.onload = () => resolve(String(reader.result))
          reader.onerror = reject
          reader.readAsDataURL(file)
        })
        const image = new Image()
        image.src = result
        await image.decode()
        if (version === this.readVersion) this.$emit('update:modelValue', result)
      } catch {
        this.error = 'Не удалось прочитать изображение.'
      }
      input.value = ''
    },
  },
})
</script>
