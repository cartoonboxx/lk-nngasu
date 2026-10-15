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

    <section :class="style.section" aria-labelledby="inputs-title">
      <div :class="style.sectionHeading">
        <h2 id="inputs-title">Поля ввода</h2>
      </div>

      <div :class="style.showcase">
        <article :class="style.group">
          <h3>Варианты</h3>
          <div :class="style.row">
            <Input
              v-for="variant in inputVariants"
              :key="variant.value"
              :variant="variant.value"
              :placeholder="variant.label"
              :model-value="variant.value === 'default' ? '' : variant.label"
            />
          </div>
        </article>

        <article :class="style.group">
          <h3>Размеры</h3>
          <div :class="style.row">
            <Input
              v-for="size in inputSizes"
              :key="size.value"
              :size="size.value"
              :placeholder="size.label"
            />
          </div>
        </article>

        <article :class="style.group">
          <h3>Textarea</h3>
          <div :class="style.row">
            <Input
              v-model="textareaValue"
              multiline
              full-width
              :base-height="96"
              variant="filled"
              placeholder="Описание"
            />
          </div>
        </article>

        <article :class="style.group">
          <h3>Prefix slot</h3>
          <div :class="style.row">
            <Input placeholder="Поиск по сайту" variant="filled">
              <template #prefix>
                <PhMagnifyingGlass :size="20" aria-hidden="true" />
              </template>
            </Input>
          </div>
        </article>

        <article :class="style.group">
          <h3>Actions</h3>
          <div :class="style.row">
            <Input placeholder="Поиск" variant="filled">
              <template #actions>
                <IconButton ariaLabel="Найти" size="sm" variant="ghost">
                  <PhAirplaneTaxiing />
                </IconButton>
              </template>
            </Input>
          </div>
        </article>
      </div>
    </section>

    <section :class="style.section" aria-labelledby="icon-buttons-title">
      <div :class="style.sectionHeading">
        <h2 id="icon-buttons-title">Кнопки-иконки</h2>
      </div>

      <div :class="style.showcase">
        <article :class="style.group">
          <h3>Варианты</h3>
          <div :class="style.row">
            <IconButton
              v-for="variant in iconButtonVariants"
              :key="variant.value"
              :ariaLabel="variant.label"
              :variant="variant.value"
            >
              <PhHouse v-if="variant.value === 'danger'" />
              <PhAirplaneTaxiing v-else />
            </IconButton>
          </div>
        </article>

        <article :class="style.group">
          <h3>Размеры</h3>
          <div :class="style.row">
            <IconButton
              v-for="size in iconButtonSizes"
              :key="size.value"
              :ariaLabel="size.label"
              :size="size.value"
            >
              <PhAirplaneTaxiing />
            </IconButton>
          </div>
        </article>

        <article :class="style.group">
          <h3>Состояния</h3>
          <div :class="style.states">
            <div v-for="variant in iconButtonVariants" :key="variant.value" :class="style.stateRow">
              <span :class="style.stateLabel">{{ variant.label }}</span>
              <IconButton
                v-for="state in iconButtonStates"
                :key="state.value"
                :ariaLabel="`${variant.label} ${state.label}`"
                :variant="variant.value"
                :state="state.value"
              >
                <PhHouse v-if="variant.value === 'danger'" />
                <PhAirplaneTaxiing v-else />
              </IconButton>
            </div>
          </div>
        </article>
      </div>
    </section>

    <section :class="style.section" aria-labelledby="badge-title">
      <div :class="style.sectionHeading">
        <h2 id="badge-title">Badge</h2>
      </div>

      <div :class="style.showcase">
        <article :class="style.group">
          <h3>Расположение</h3>
          <div :class="style.row">
            <Badge :value="3">
              <IconButton ariaLabel="Уведомления" variant="secondary">
                <PhHouse />
              </IconButton>
            </Badge>
            <Badge :value="12">
              <span :class="style.badgeTarget">Элемент</span>
            </Badge>
          </div>
        </article>
      </div>
    </section>

    <section :class="style.section" aria-labelledby="icon-tabs-title">
      <div :class="style.sectionHeading">
        <h2 id="icon-tabs-title">Табы-иконки</h2>
      </div>

      <div :class="style.showcase">
        <article :class="style.group">
          <h3>Варианты</h3>
          <div :class="style.row" role="tablist" aria-label="Варианты табов-иконок">
            <IconTab
              v-for="variant in iconTabVariants"
              :key="variant.value"
              :ariaLabel="variant.label"
              :variant="variant.value"
            >
              <PhAirplaneTaxiing />
            </IconTab>
          </div>
        </article>

        <article :class="style.group">
          <h3>Размеры</h3>
          <div :class="style.row" role="tablist" aria-label="Размеры табов-иконок">
            <IconTab
              v-for="size in iconTabSizes"
              :key="size.value"
              :ariaLabel="size.label"
              :size="size.value"
            >
              <PhAirplaneTaxiing />
            </IconTab>
          </div>
        </article>

        <article :class="style.group">
          <h3>Состояния</h3>
          <div :class="style.states" role="tablist" aria-label="Состояния табов-иконок">
            <div v-for="variant in iconTabVariants" :key="variant.value" :class="style.stateRow">
              <span :class="style.stateLabel">{{ variant.label }}</span>
              <IconTab
                v-for="state in iconTabStates"
                :key="state.value"
                :ariaLabel="`${variant.label} ${state.label}`"
                :variant="variant.value"
                :state="state.value"
              >
                <PhAirplaneTaxiing />
              </IconTab>
            </div>
          </div>
        </article>
      </div>
    </section>

    <section :class="style.section" aria-labelledby="tabs-title">
      <div :class="style.sectionHeading">
        <h2 id="tabs-title">Табы</h2>
      </div>

      <div :class="style.showcase">
        <article :class="style.group">
          <h3>Внешний вид</h3>
          <div :class="style.tabColumn" role="tablist" aria-label="Пример таба">
            <Tab text="Главная" :icon="tabIcon" active>
              <template #actions>
                <span :class="style.tabBadge">2</span>
              </template>
            </Tab>
          </div>
        </article>

        <article :class="style.group">
          <h3>Состав</h3>
          <div :class="style.tabColumn" role="tablist" aria-label="Состав табов">
            <Tab text="Только текст" />
            <Tab text="Изображение без actions" :image-src="logoImage" />
            <Tab text="Недоступный таб" disabled />
          </div>
        </article>

        <article :class="style.group">
          <h3>Активное состояние</h3>
          <div :class="style.tabColumn" role="tablist" aria-label="Активные табы">
            <Tab text="Активный таб" :icon="tabIcon" active />
          </div>
        </article>
      </div>
    </section>
  </main>
</template>

<script lang="ts">
import { PhAirplaneTaxiing, PhHouse, PhMagnifyingGlass } from '@phosphor-icons/vue'
import { defineComponent, markRaw } from 'vue'

import Badge from '@/components/UI/Badge/Badge.vue'
import Button from '@/components/UI/buttons/Button/Button.vue'
import IconButton from '@/components/UI/buttons/IconButton/IconButton.vue'
import Input from '@/components/UI/inputs/Input/Input.vue'
import IconTab from '@/components/UI/tabs/IconTab/IconTab.vue'
import Tab from '@/components/UI/tabs/Tab/Tab.vue'
import logoImage from '$images/logo.svg'
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

const inputVariants: Array<{ value: InputVariant; label: string }> = [
  { value: 'default', label: 'Default' },
  { value: 'filled', label: 'Filled' },
  { value: 'error', label: 'Error' },
  { value: 'disabled', label: 'Disabled' },
]

const inputSizes: Array<{ value: InputSize; label: string }> = [
  { value: 'sm', label: 'Small' },
  { value: 'md', label: 'Medium' },
  { value: 'lg', label: 'Large' },
]

const iconButtonVariants: Array<{ value: IconButtonVariant; label: string }> = [
  { value: 'primary', label: 'Primary' },
  { value: 'secondary', label: 'Secondary' },
  { value: 'ghost', label: 'Ghost' },
  { value: 'danger', label: 'Danger' },
]

const iconButtonSizes: Array<{ value: IconButtonSize; label: string }> = [
  { value: 'sm', label: 'Small' },
  { value: 'md', label: 'Medium' },
  { value: 'lg', label: 'Large' },
]

const iconButtonStates: Array<{ value: IconButtonState; label: string }> = [
  { value: 'default', label: 'Default' },
  { value: 'hover', label: 'Hover' },
  { value: 'active', label: 'Active' },
  { value: 'disabled', label: 'Disabled' },
]

const iconTabVariants: Array<{ value: IconTabVariant; label: string }> = [
  { value: 'default', label: 'Default' },
  { value: 'selected', label: 'Selected' },
]

const iconTabSizes: Array<{ value: IconTabSize; label: string }> = [
  { value: 'sm', label: 'Small' },
  { value: 'md', label: 'Medium' },
  { value: 'lg', label: 'Large' },
]

const iconTabStates: Array<{ value: IconTabState; label: string }> = [
  { value: 'default', label: 'Default' },
  { value: 'hover', label: 'Hover' },
  { value: 'active', label: 'Active' },
  { value: 'disabled', label: 'Disabled' },
]

export default defineComponent({
  name: 'UiKitPage',
  components: {
    Badge,
    Button,
    PhHouse,
    PhAirplaneTaxiing,
    PhMagnifyingGlass,
    IconButton,
    IconTab,
    Input,
    Tab,
  },
  data() {
    return {
      style,
      variants,
      sizes,
      states,
      inputVariants,
      inputSizes,
      iconButtonVariants,
      iconButtonSizes,
      iconButtonStates,
      iconTabVariants,
      iconTabSizes,
      iconTabStates,
      tabIcon: markRaw(PhHouse),
      logoImage,
      textareaValue:
        'Textarea использует базовую высоту 96px и расширяется по высоте, когда текста становится больше.',
    }
  },
})
</script>
