<template>
  <main :class="style.page">
    <header :class="style.header">
      <p :class="style.caption">UI Kit</p>
      <h1>Компоненты интерфейса</h1>
    </header>

    <section :class="style.section">
      <div :class="style.sectionHeading">
        <h2>Tooltip</h2>
      </div>
      <div :class="style.showcase">
        <article :class="style.group">
          <h3>Расположение</h3>
          <div :class="style.row">
            <Tooltip
              v-for="position in tooltipPositions"
              :key="position"
              :position="position"
              :text="`Подсказка: ${position}`"
            >
              <Button variant="secondary">{{ position }}</Button>
            </Tooltip>
            <Tooltip position="right">
              <span :class="style.badgeTarget">Разметка</span>
              <template #content>Подсказка с <strong>разметкой</strong></template>
            </Tooltip>
          </div>
        </article>
      </div>
    </section>

    <section :class="style.section">
      <div :class="style.sectionHeading">
        <h2>Кнопки</h2>
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

    <section :class="style.section">
      <div :class="style.sectionHeading">
        <h2>Медиафайлы</h2>
      </div>
      <div :class="style.showcase">
        <article :class="style.group">
          <h3>Скачивание материалов</h3>
          <div :class="style.mediaList">
            <MediaButton title="Лекция 1" meta="Конспект · MD" @click="downloadMediaExample">
              <template #icon
                ><PhFileText :class="style.mediaError" :size="30" weight="fill"
              /></template>
            </MediaButton>
            <MediaButton
              title="План темы"
              meta="Учебный материал · MD"
              @click="downloadMediaExample"
            >
              <template #icon
                ><PhPresentationChart :class="style.mediaWarning" :size="30" weight="fill"
              /></template>
            </MediaButton>
            <MediaButton
              title="Практическая работа №1"
              meta="Задание · MD"
              @click="downloadMediaExample"
            >
              <template #icon><PhFileCode :size="30" weight="fill" /></template>
            </MediaButton>
            <MediaButton title="Недоступный материал" meta="Скачивание недоступно" disabled>
              <template #icon><PhFileText :size="30" /></template>
            </MediaButton>
          </div>
        </article>
        <article :class="style.group">
          <h3>Readonly — прикреплённые файлы</h3>
          <div :class="style.mediaList">
            <MediaButton v-if="mediaAttachmentVisible" title="main.pl" meta="4 КБ" readonly>
              <template #icon><PhFileCode :size="34" weight="duotone" /></template>
              <template #actions>
                <IconButton variant="ghost" size="sm" @click="mediaAttachmentVisible = false"
                  ><PhX :size="20"
                /></IconButton>
              </template>
            </MediaButton>
            <Button v-else variant="secondary" @click="mediaAttachmentVisible = true"
              >Вернуть файл main.pl</Button
            >
            <MediaButton title="Отчёт.pdf" meta="1,2 МБ" readonly>
              <template #icon
                ><PhFilePdf :class="style.mediaError" :size="34" weight="duotone"
              /></template>
            </MediaButton>
            <MediaButton readonly>
              <template #icon><PhFileText :size="34" weight="duotone" /></template>
              <span>Произвольное содержимое через слот</span>
              <small>Дополнительная информация о файле</small>
            </MediaButton>
          </div>
        </article>
      </div>
    </section>

    <section :class="style.section">
      <div :class="style.sectionHeading">
        <h2>Поля ввода</h2>
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
                <PhMagnifyingGlass :size="20" />
              </template>
            </Input>
          </div>
        </article>

        <article :class="style.group">
          <h3>Actions</h3>
          <div :class="style.row">
            <Input placeholder="Поиск" variant="filled">
              <template #actions>
                <IconButton size="sm" variant="ghost">
                  <PhAirplaneTaxiing />
                </IconButton>
              </template>
            </Input>
          </div>
        </article>
      </div>
    </section>

    <section :class="style.section">
      <div :class="style.sectionHeading">
        <h2>Select</h2>
      </div>
      <div :class="style.showcase">
        <article :class="style.group">
          <h3>Выбор значения</h3>
          <div :class="style.row">
            <Select
              v-model="selectedSubject"
              :options="selectOptions"
              placeholder="Выберите предмет"
            />
            <span>Выбрано: {{ selectedSubject || 'ничего' }}</span>
          </div>
        </article>
        <article :class="style.group">
          <h3>Состояния</h3>
          <div :class="style.row">
            <Select :options="selectOptions" model-value="math" />
            <Select :options="selectOptions" model-value="math" disabled />
            <Select :options="[]" placeholder="Пустой список" />
          </div>
        </article>
        <article :class="style.group">
          <h3>Размеры</h3>
          <div :class="style.row">
            <Select
              v-for="size in inputSizes"
              :key="size.value"
              v-model="selectedSubject"
              :options="selectOptions"
              :size="size.value"
              :placeholder="size.label"
            />
          </div>
        </article>
        <article :class="style.group">
          <h3>На всю ширину</h3>
          <Select v-model="selectedSubject" :options="selectOptions" full-width />
        </article>
      </div>
    </section>

    <section :class="style.section">
      <div :class="style.sectionHeading">
        <h2>FloatingPopup</h2>
      </div>
      <div :class="style.showcase">
        <article :class="style.group">
          <h3>Вкладки в виде списка</h3>
          <FloatingPopup
            v-model="popupTab"
            :items="switchableTabs"
            trigger-text="Учебная информация"
          >
            <template #schedule>Расписание занятий на текущую неделю.</template>
            <template #subjects>Список предметов текущего семестра.</template>
            <template #grades>Оценки и результаты аттестации.</template>
          </FloatingPopup>
        </article>
      </div>
    </section>

    <section :class="style.section">
      <div :class="style.sectionHeading">
        <h2>Кнопки-иконки</h2>
      </div>

      <div :class="style.showcase">
        <article :class="style.group">
          <h3>Варианты</h3>
          <div :class="style.row">
            <IconButton
              v-for="variant in iconButtonVariants"
              :key="variant.value"
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
            <IconButton v-for="size in iconButtonSizes" :key="size.value" :size="size.value">
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

    <section :class="style.section">
      <div :class="style.sectionHeading">
        <h2>Badge</h2>
      </div>

      <div :class="style.showcase">
        <article :class="style.group">
          <h3>Расположение</h3>
          <div :class="style.row">
            <Badge :value="3">
              <IconButton variant="secondary">
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

    <section :class="style.section">
      <div :class="style.sectionHeading">
        <h2>Табы-иконки</h2>
      </div>

      <div :class="style.showcase">
        <article :class="style.group">
          <h3>Варианты</h3>
          <div :class="style.row">
            <IconTab
              v-for="variant in iconTabVariants"
              :key="variant.value"
              :variant="variant.value"
            >
              <PhAirplaneTaxiing />
            </IconTab>
          </div>
        </article>

        <article :class="style.group">
          <h3>Размеры</h3>
          <div :class="style.row">
            <IconTab v-for="size in iconTabSizes" :key="size.value" :size="size.value">
              <PhAirplaneTaxiing />
            </IconTab>
          </div>
        </article>

        <article :class="style.group">
          <h3>Состояния</h3>
          <div :class="style.states">
            <div v-for="variant in iconTabVariants" :key="variant.value" :class="style.stateRow">
              <span :class="style.stateLabel">{{ variant.label }}</span>
              <IconTab
                v-for="state in iconTabStates"
                :key="state.value"
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

    <section :class="style.section">
      <div :class="style.sectionHeading">
        <h2>Табы</h2>
      </div>

      <div :class="style.showcase">
        <article :class="style.group">
          <h3>BaseTab — общая геометрия</h3>
          <div :class="style.tabColumn">
            <BaseTab text="Базовый таб" :icon="tabIcon" />
          </div>
        </article>

        <article :class="style.group">
          <h3>NavTab — навигация</h3>
          <div :class="style.tabColumn">
            <NavTab
              v-for="item in switchableTabs"
              :key="item.value"
              :text="item.label"
              :active="item.value === activeTab"
              :disabled="item.disabled"
              :icon="tabIcon"
              @select="activeTab = item.value"
            />
          </div>
        </article>

        <article :class="style.group">
          <h3>StyleTab — выбор в списке</h3>
          <div :class="style.tabColumn">
            <StyleTab
              v-for="item in selectOptions"
              :key="item.value"
              :text="item.label"
              :active="item.value === selectedSubject"
              :disabled="item.disabled"
              @select="selectedSubject = item.value"
            />
          </div>
        </article>

        <article :class="style.group">
          <h3>StyleTab — дополнительные действия</h3>
          <div :class="style.tabColumn">
            <StyleTab text="Главная" :icon="tabIcon" active>
              <template #actions>
                <span :class="style.tabBadge">2</span>
              </template>
            </StyleTab>
          </div>
        </article>

        <article :class="style.group">
          <h3>StyleTab — состав</h3>
          <div :class="style.tabColumn">
            <StyleTab text="Только текст" />
            <StyleTab text="Изображение без actions" :image-src="logoImage" />
            <StyleTab text="Недоступный таб" disabled />
          </div>
        </article>

        <article :class="style.group">
          <h3>StyleTab — выбранное состояние</h3>
          <div :class="style.tabColumn">
            <StyleTab text="Активный таб" :icon="tabIcon" active />
          </div>
        </article>
      </div>
    </section>
    <section :class="style.section">
      <div :class="style.sectionHeading">
        <h2>Модальное окно</h2>
      </div>
      <div :class="style.showcase">
        <article :class="style.group">
          <h3>Произвольное содержимое</h3>
          <div :class="style.row">
            <Button @click="modalOpen = true">Открыть модальное окно</Button>
          </div>
        </article>
      </div>
      <Modal v-model="modalOpen" v-slot="{ close }">
        <div :class="style.modalActions">
          <div :class="style.header">
            <h3>Обратная связь</h3>
            <p>Напишите сообщение, и мы постараемся помочь.</p>
          </div>
          <IconButton variant="ghost" @click="close">
            <PhX :size="20" />
          </IconButton>
        </div>
        <Input v-model="modalMessage" multiline full-width placeholder="Ваше сообщение" />
      </Modal>
    </section>
  </main>
</template>

<script lang="ts">
import {
  PhAirplaneTaxiing,
  PhHouse,
  PhMagnifyingGlass,
  PhX,
  PhFileText,
  PhFileCode,
  PhFilePdf,
  PhPresentationChart,
} from '@phosphor-icons/vue'
import { defineComponent, markRaw } from 'vue'

import Tooltip, { type TooltipPosition } from '@/components/UI/Tooltip/Tooltip.vue'
import Badge from '@/components/UI/Badge/Badge.vue'
import Button from '@/components/UI/buttons/Button/Button.vue'
import MediaButton from '@/components/UI/buttons/MediaButton/MediaButton.vue'
import IconButton from '@/components/UI/buttons/IconButton/IconButton.vue'
import Input from '@/components/UI/inputs/Input/Input.vue'
import Select from '@/components/UI/inputs/Select/Select.vue'
import Modal from '@/components/UI/modal/BaseModal/Modal.vue'
import FloatingPopup from '@/components/UI/popup/FloatingPopup/FloatingPopup.vue'
import IconTab from '@/components/UI/tabs/IconTab/IconTab.vue'
import StyleTab from '@/components/UI/tabs/StyleTab/StyleTab.vue'
import BaseTab from '@/components/UI/tabs/BaseTab/BaseTab.vue'
import NavTab from '@/components/UI/tabs/NavTab/NavTab.vue'
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
    Tooltip,
    MediaButton,
    PhFileText,
    PhFileCode,
    PhFilePdf,
    PhPresentationChart,
    Badge,
    Button,
    PhHouse,
    PhX,
    PhAirplaneTaxiing,
    PhMagnifyingGlass,
    IconButton,
    IconTab,
    Input,
    Select,
    FloatingPopup,
    Modal,
    StyleTab,
    BaseTab,
    NavTab,
  },
  data() {
    return {
      style,
      tooltipPositions: ['top', 'bottom', 'left', 'right'] as TooltipPosition[],
      mediaAttachmentVisible: true,
      modalOpen: false,
      modalMessage: '',
      activeTab: 'schedule',
      popupTab: 'schedule',
      selectedSubject: '',
      selectOptions: [
        { value: 'math', label: 'Математика' },
        { value: 'physics', label: 'Физика' },
        { value: 'drawing', label: 'Начертательная геометрия' },
        { value: 'archive', label: 'Архивный предмет', disabled: true },
      ],
      switchableTabs: [
        { value: 'schedule', label: 'Расписание' },
        { value: 'subjects', label: 'Предметы' },
        { value: 'grades', label: 'Оценки' },
        { value: 'archive', label: 'Архив', disabled: true },
      ],
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
  methods: {
    downloadMediaExample() {
      const url = URL.createObjectURL(
        new Blob(['# Учебный материал\n\nПример файла из UI Kit.'], {
          type: 'text/markdown;charset=utf-8',
        }),
      )
      const link = document.createElement('a')
      link.href = url
      link.download = 'material.md'
      link.click()
      setTimeout(() => URL.revokeObjectURL(url), 1000)
    },
  },
})
</script>
