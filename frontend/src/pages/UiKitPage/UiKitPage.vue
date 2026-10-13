<template>
  <main :class="style.page">
    <header :class="style.header">
      <p :class="style.caption">UI Kit</p>
      <h1>Компоненты интерфейса</h1>
    </header>

    <section :class="style.section" aria-labelledby="buttons-title">
      <div :class="style.sectionHeading">
        <h2 id="buttons-title">Кнопки</h2>
      </div>

      <div :class="style.showcase">
        <article :class="style.group">
          <h3>Варианты</h3>
          <div :class="style.row">
            <Button v-for="variant in variants" :key="variant.value" :variant="variant.value">
              {{ variant.label }}
            </Button>
          </div>
        </article>

        <article :class="style.group">
          <h3>Размеры</h3>
          <div :class="style.row">
            <Button v-for="size in sizes" :key="size.value" :size="size.value">
              {{ size.label }}
            </Button>
          </div>
        </article>

        <article :class="style.group">
          <h3>Состояния</h3>
          <div :class="style.states">
            <div v-for="variant in variants" :key="variant.value" :class="style.stateRow">
              <span :class="style.stateLabel">{{ variant.label }}</span>
              <Button
                v-for="state in states"
                :key="state.value"
                :variant="variant.value"
                :state="state.value"
              >
                {{ state.label }}
              </Button>
            </div>
          </div>
        </article>
      </div>
    </section>
  </main>
</template>

<script lang="ts">
import { defineComponent } from 'vue'

import Button from '@/components/Button/Button.vue'
import style from './UiKitPage.module.scss'

const variants: Array<{ value: ButtonVariant; label: string }> = [
  { value: 'primary', label: 'Primary' },
  { value: 'secondary', label: 'Secondary' },
  { value: 'ghost', label: 'Ghost' },
  { value: 'danger', label: 'Danger' },
]

const sizes: Array<{ value: ButtonSize; label: string }> = [
  { value: 'sm', label: 'Small' },
  { value: 'md', label: 'Medium' },
  { value: 'lg', label: 'Large' },
]

const states: Array<{ value: ButtonState; label: string }> = [
  { value: 'default', label: 'Default' },
  { value: 'hover', label: 'Hover' },
  { value: 'active', label: 'Active' },
  { value: 'disabled', label: 'Disabled' },
]

export default defineComponent({
  name: 'UiKitPage',
  components: {
    Button,
  },
  data() {
    return {
      style,
      variants,
      sizes,
      states,
    }
  },
})
</script>
