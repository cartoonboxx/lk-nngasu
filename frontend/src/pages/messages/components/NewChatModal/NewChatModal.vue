<template>
  <Modal
    :model-value="modelValue"
    :width="440"
    @update:model-value="!opening && $emit('update:modelValue', $event)"
  >
    <header :class="style.modalHeader">
      <h2>Новый диалог</h2>
      <Button variant="ghost" :disabled="opening" @click="$emit('update:modelValue', false)"
        >Закрыть</Button
      >
    </header>
    <TextInput
      v-model="search"
      label="Поиск контактов"
      placeholder="Имя пользователя"
      type="search"
      full-width
      :maxlength="200"
    />
    <p v-if="error" :class="style.contactsEmpty">
      {{ error }} <button @click="loadContacts()">Повторить</button>
    </p>
    <PageLoader v-if="loading" />
    <p v-else-if="!error && !contacts.length" :class="style.contactsEmpty">Контакты не найдены.</p>
    <ul :class="style.contacts">
      <li v-for="contact in contacts" :key="contact.id">
        <Button
          :class="style.contact"
          variant="ghost"
          :disabled="opening"
          @click="openDirect(contact.id)"
        >
          <span :class="style.contactCopy"
            ><strong>{{ contact.name }}</strong
            ><span>{{ contact.description }}</span></span
          >
        </Button>
      </li>
    </ul>
    <Button
      v-if="contacts.length < total"
      variant="ghost"
      :disabled="loading || opening"
      @click="loadContacts(true)"
      >Ещё контакты</Button
    >
    <Button v-if="auth.user?.student" variant="secondary" :disabled="opening" @click="openGroup"
      >Чат группы {{ auth.user.student.group_name }}</Button
    >
  </Modal>
</template>
<script lang="ts">
import { apiNotifications, type ApiPage } from '@/api/client'
import type { Contact } from '@/api/cabinet.types'
import { defineComponent } from 'vue'
import Modal from '@/components/UI/modal/BaseModal/Modal.vue'
import Button from '@/components/UI/buttons/Button/Button.vue'
import TextInput from '@/components/UI/inputs/TextInput/TextInput.vue'
import { useAuthStore } from '@/stores/auth'
import style from './NewChatModal.module.scss'
import PageLoader from '@/components/PageLoader/PageLoader.vue'
export default defineComponent({
  name: 'NewChatModal',
  components: { Modal, Button, TextInput, PageLoader },
  props: { modelValue: Boolean },
  emits: ['update:modelValue', 'select'],
  computed: {
    auth() {
      return useAuthStore()
    },
  },
  data() {
    return {
      style,
      search: '',
      contacts: [] as Contact[],
      total: 0,
      loading: false,
      error: '',
      opening: false,
      version: 0,
      timer: 0,
    }
  },
  watch: {
    modelValue: {
      immediate: true,
      handler(open: boolean) {
        this.version++
        window.clearTimeout(this.timer)
        this.contacts = []
        this.search = ''
        if (open) void this.loadContacts()
      },
    },
    search() {
      this.version++
      window.clearTimeout(this.timer)
      if (this.modelValue) this.timer = window.setTimeout(() => void this.loadContacts(), 250)
    },
  },
  beforeUnmount() {
    this.version++
    window.clearTimeout(this.timer)
  },
  methods: {
    async loadContacts(append = false) {
      const version = ++this.version
      this.loading = true
      this.error = ''
      if (!append) this.contacts = []
      try {
        const query = new URLSearchParams({
          q: this.search.trim(),
          limit: '50',
          offset: String(append ? this.contacts.length : 0),
        })
        const page = await this.auth.request<ApiPage<Contact>>(`/contacts?${query}`, {
          paginated: true,
        })
        if (version !== this.version || !this.modelValue) return
        this.contacts = append ? [...this.contacts, ...page.data] : page.data
        this.total = page.meta.total
      } catch (error) {
        if (version === this.version)
          this.error = error instanceof Error ? error.message : 'Не удалось загрузить контакты.'
      } finally {
        if (version === this.version) this.loading = false
      }
    },
    async openDirect(userId: number) {
      if (this.opening) return
      this.opening = true
      try {
        const id = await this.$messageStore.openDirect(userId)
        this.$emit('select', id)
      } catch (error) {
        apiNotifications.onError?.(error)
      } finally {
        this.opening = false
      }
    },
    async openGroup() {
      if (this.opening || !this.auth.user?.student) return
      this.opening = true
      try {
        const id = await this.$messageStore.openGroup(this.auth.user.student.group_id)
        this.$emit('select', id)
      } catch (error) {
        apiNotifications.onError?.(error)
      } finally {
        this.opening = false
      }
    },
  },
})
</script>
