<template>
  <Modal
    :model-value="modelValue"
    :width="440"
    @update:model-value="$emit('update:modelValue', $event)"
  >
    <header :class="style.header">
      <div>
        <h2>Участники группы</h2>
        <p>{{ chatName }}</p>
      </div>
      <IconButton variant="ghost" title="Закрыть" @click="$emit('update:modelValue', false)">
        <PhX :size="20" />
      </IconButton>
    </header>
    <PageLoader v-if="loading" />
    <div v-else-if="error" :class="style.feedback">
      <p>{{ error }}</p>
      <Button variant="secondary" @click="loadParticipants">Повторить</Button>
    </div>
    <p v-else-if="!participants.length" :class="style.feedback">Участников пока нет.</p>
    <ul v-else :class="style.participants">
      <li v-for="person in participants" :key="person.id">
        <button
          type="button"
          :class="style.participant"
          :disabled="person.id === currentUserId || openingId !== null"
          @click="openDirect(person.id)"
        >
          <ChatAvatar
            :initials="
              person.name
                .split(/\s+/)
                .filter(Boolean)
                .slice(0, 2)
                .map((part) => part[0])
                .join('')
            "
          />
          <span :class="style.personName">{{ person.name }}</span>
          <span :class="style.actionText">
            {{
              person.id === currentUserId
                ? 'Вы'
                : openingId === person.id
                  ? 'Открываем…'
                  : 'Написать'
            }}
          </span>
        </button>
      </li>
    </ul>
  </Modal>
</template>

<script lang="ts">
import { defineComponent } from 'vue'
import { PhX } from '@phosphor-icons/vue'
import { apiNotifications } from '@/api/client'
import { useAuthStore } from '@/stores/auth'
import Modal from '@/components/UI/modal/BaseModal/Modal.vue'
import Button from '@/components/UI/buttons/Button/Button.vue'
import IconButton from '@/components/UI/buttons/IconButton/IconButton.vue'
import ChatAvatar from '../ChatAvatar/ChatAvatar.vue'
import style from './ChatParticipantsModal.module.scss'
import PageLoader from '@/components/PageLoader/PageLoader.vue'
export default defineComponent({
  name: 'ChatParticipantsModal',
  components: { Modal, Button, IconButton, ChatAvatar, PhX, PageLoader },
  props: {
    modelValue: Boolean,
    chatId: { type: String, required: true },
    chatName: { type: String, required: true },
  },
  emits: ['update:modelValue', 'select'],
  data() {
    return {
      style,
      participants: [] as { id: number; name: string }[],
      loading: false,
      error: '',
      openingId: null as number | null,
      version: 0,
    }
  },
  computed: {
    currentUserId() {
      return useAuthStore().user?.id
    },
  },
  watch: {
    modelValue(open: boolean) {
      this.version++
      if (open) void this.loadParticipants()
    },
    chatId() {
      this.version++
      this.participants = []
    },
  },
  beforeUnmount() {
    this.version++
  },
  methods: {
    async loadParticipants() {
      const version = ++this.version
      this.loading = true
      this.error = ''
      this.participants = []
      try {
        const people = await useAuthStore().request<{ id: number; name: string }[]>(
          `/chats/${this.chatId}/participants`,
          { notifyOnError: false },
        )
        if (version === this.version) this.participants = people
      } catch (error) {
        if (version === this.version)
          this.error = error instanceof Error ? error.message : 'Не удалось загрузить участников.'
      } finally {
        if (version === this.version) this.loading = false
      }
    },
    async openDirect(userId: number) {
      if (this.openingId !== null || userId === this.currentUserId) return
      const version = this.version
      this.openingId = userId
      try {
        const id = await this.$messageStore.openDirect(userId)
        if (version !== this.version) return
        this.$emit('update:modelValue', false)
        this.$emit('select', id)
      } catch (error) {
        if (version === this.version) apiNotifications.onError?.(error)
      } finally {
        this.openingId = null
      }
    },
  },
})
</script>
