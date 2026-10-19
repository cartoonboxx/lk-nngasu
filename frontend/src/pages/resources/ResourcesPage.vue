<template>
    <section :class="style.header">
      <h1>Ресурсы и сервисы</h1>
      <p>Учебные платформы, боты и контакты университета</p>
    </section>

    <section :class="style.pageInput">
      <StyleInput
        v-model="searchQuery"
        placeholder="Поиск ресурса или службы"
        :full-width="true">
        
        <template #prefix>
          <PhMagnifyingGlass :size="20" aria-hidden="true" />
        </template>
      </StyleInput>
    </section>

    <section :class="style.section">
      <div :class="style.sectionHeading">
        <h2>Сервисы университета</h2>
        <p>Открываются в новой вкладке</p>
      </div>

      <div :class="style.servicesGrid">
        <article
          v-for="service in filteredServices"
          :key="service.label"
          :class="style.serviceCard">

          <div :class="style.cardContent">
            <h3>{{ service.label }}</h3>
            <p>{{ service.description }}</p>

            <a
              :href="service.url"
              target="_blank"
              rel="noopener noreferrer"
              :class="style.cardLink">

              {{ service.linkText }}
              <PhArrowSquareOut :size="18" />
            </a>
          </div>

          <div :class="style.serviceIcon">
            <component :is="service.icon" :size="56" weight="fill" />
          </div>
        </article>
      </div>
    </section>

    <section :class="style.section">
      <div :class="style.sectionHeading">
        <h2>Контакты университета</h2>
      </div>

      <div :class="style.contactsGrid">
        <article
          v-for="contact in filteredContacts"
          :key="contact.label"
          :class="style.contactCard">

          <div :class="style.contactIcon">
            <component :is="contact.icon" :size="28" />
          </div>

          <div :class="style.contactContent">
            <h3>{{ contact.label }}</h3>
            <p>{{ contact.description }}</p>

            <a :href="contact.url" :class="style.cardLink">
              {{ contact.linkText }}
              <PhArrowSquareOut :size="18" />
            </a>
          </div>
        </article>
      </div>

    </section>
</template>

<script lang="ts">
import { defineComponent } from 'vue'
import style from './ResourcesPage.module.scss'
import VK_logo from '$images/vk_logo.svg'


import {
  PhMagnifyingGlass,
  PhArrowSquareOut,
  PhGraduationCap,
  PhPlayCircle,
  PhTelegramLogo,
  PhChatCircle,
  PhUsersThree,
  PhFileText,
  PhHeadset,
} from '@phosphor-icons/vue'

import StyleInput from '@/components/UI/inputs/Input/Input.vue'

interface Service {
  label: string
  description: string
  domain: string
  url: string
  linkText: string
  icon: typeof PhGraduationCap
}

interface Contact {
  label: string
  description: string
  url: string
  linkText: string
  icon: typeof PhUsersThree
}

export default defineComponent({
  name: 'ResourcesPage',

  components: {
    StyleInput,
    PhMagnifyingGlass,
    PhArrowSquareOut,
  },

  data() {
    return {
      style,
      searchQuery: '',

      services: [
        {
          label: 'Moodle LMS',
          description: 'Учебные материалы, задания и курсы в системе электронного обучения.',
          domain: 'lms.nggasu.ru',
          url: 'https://lms.nngasu.ru/',
          linkText: 'Открыть Moodle',
          icon: PhGraduationCap,
        },
        {
          label: 'Видеолекции',
          description: 'Записи лекций и видеоматериалы по дисциплинам.',
          domain: '',
          url: '#',
          linkText: 'Смотреть лекции',
          icon: PhPlayCircle,
        },
        {
          label: 'ВК',
          description: 'Доступ к сервисам личного кабинета через ВК.',
          domain: '',
          url: 'https://vk.ru/nngasu',
          linkText: 'Открыть в ВК',
          icon: VK_logo,
        },
        {
          label: 'MAX-бот',
          description: 'Взаимодействие с личным кабинетом в мессенджере MAX.',
          domain: '',
          url: 'https://max.ru',
          linkText: 'Открыть в MAX',
          icon: PhChatCircle,
        },
      ] as Service[],

      contacts: [
        {
          label: 'Учебный отдел',
          description: 'Справки и вопросы по учебному процессу.',
          url: '#',
          linkText: 'Посмотреть контакты',
          icon: PhUsersThree,
        },
        {
          label: 'Деканат',
          description: 'Организационные вопросы и документы.',
          url: '#',
          linkText: 'Посмотреть контакты',
          icon: PhFileText,
        },
        {
          label: 'Техническая поддержка',
          description: 'Помощь с личным кабинетом и сервисами.',
          url: '#',
          linkText: 'Написать в поддержку',
          icon: PhHeadset,
        },
      ] as Contact[],
    }
  },

  computed: {
    filteredServices(): Service[] {
      const query = this.searchQuery.trim().toLowerCase()

      return this.services.filter((service) =>
        `${service.label} ${service.description} ${service.domain}`
          .toLowerCase()
          .includes(query),
      )
    },

    filteredContacts(): Contact[] {
      const query = this.searchQuery.trim().toLowerCase()

      return this.contacts.filter((contact) =>
        `${contact.label} ${contact.description}`
          .toLowerCase()
          .includes(query),
      )
    },
  },
})
</script>
