<template>
  <section :class="style.header">
    <h1>Ресурсы и сервисы</h1>
    <p>Учебные платформы, боты и контакты университета</p>
  </section>

  <section :class="style.pageInput">
    <StyleInput v-model="searchQuery" placeholder="Поиск ресурса или службы" :full-width="true">
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
      <ServiceCard
        v-for="service in filteredServices"
        :key="service.label"
        :label="service.label"
        :description="service.description"
        :url="service.url"
        :link-text="service.linkText"
        :icon="service.icon"
      />
    </div>
  </section>

  <section :class="style.section">
    <div :class="style.sectionHeading">
      <h2>Контакты университета</h2>
    </div>

    <div :class="style.contactsGrid">
      <ServiceCard
        v-for="contact in filteredContacts"
        :key="contact.label"
        :label="contact.label"
        :description="contact.description"
        :url="contact.url"
        :link-text="contact.linkText"
        :icon="contact.icon"
        icon-position="left"
        variant="contact"
      />
    </div>
  </section>
</template>

<script lang="ts">
import { defineComponent } from 'vue'
import type { Component } from 'vue'

import style from './ResourcesPage.module.scss'

import VK_logo from '$images/vk-logo.svg'
import Max_logo from '$images/MAX.svg'

import ServiceCard from './components/ServiceCard/ServiceCard.vue'
import StyleInput from '@/components/UI/inputs/Input/Input.vue'

import {
  PhMagnifyingGlass,
  PhGraduationCap,
  PhPlayCircle,
  PhUsersThree,
  PhFileText,
  PhHeadset,
} from '@phosphor-icons/vue'

type Service = {
  label: string
  description: string
  domain: string
  url: string
  linkText: string
  icon: string | Component
}

type Contact = {
  label: string
  description: string
  url: string
  linkText: string
  icon: Component
}

export default defineComponent({
  name: 'ResourcesPage',

  components: {
    StyleInput,
    ServiceCard,
    PhMagnifyingGlass,
  },

  data() {
    return {
      style,
      searchQuery: '',

      services: [
        {
          label: 'Moodle LMS',
          description: 'Учебные материалы, задания и курсы в системе электронного обучения.',
          domain: 'lms.nngasu.ru',
          url: 'https://lms.nngasu.ru/',
          linkText: 'Открыть Moodle',
          icon: PhGraduationCap,
        },
        {
          label: 'Видеолекции',
          description: 'Записи лекций и видеоматериалы по дисциплинам.',
          domain: 'видеолекции',
          url: '#',
          linkText: 'Смотреть лекции',
          icon: PhPlayCircle,
        },
        {
          label: 'ВК',
          description: 'Доступ к сервисам личного кабинета через ВК.',
          domain: 'vk.ru',
          url: 'https://vk.ru/nngasu',
          linkText: 'Открыть в ВК',
          icon: VK_logo,
        },
        {
          label: 'MAX-бот',
          description: 'Взаимодействие с личным кабинетом в мессенджере MAX.',
          domain: 'max.ru',
          url: 'https://max.ru',
          linkText: 'Открыть в MAX',
          icon: Max_logo,
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
        `${service.label} ${service.description} ${service.domain}`.toLowerCase().includes(query),
      )
    },

    filteredContacts(): Contact[] {
      const query = this.searchQuery.trim().toLowerCase()

      return this.contacts.filter((contact) =>
        `${contact.label} ${contact.description}`.toLowerCase().includes(query),
      )
    },
  },
})
</script>
