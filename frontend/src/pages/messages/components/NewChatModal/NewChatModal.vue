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
    <form :class="style.composeForm" @submit.prevent="openDirect">
      <p>Введите ID пользователя, которому хотите написать.</p>
      <TextInput v-model="userId" label="ID пользователя" inputmode="numeric" full-width required />
      <Button type="submit" :loading="opening" :disabled="opening || !/^[1-9]\d*$/.test(userId)"
        >Открыть диалог</Button
      >
    </form>
    <Button v-if="auth.user?.student" variant="secondary" :disabled="opening" @click="openGroup"
      >Чат группы {{ auth.user.student.group_name }}</Button
    >
    <p v-if="error" role="alert">{{ error }}</p>
  </Modal>
</template>
<script lang="ts">
import { defineComponent } from 'vue'
import Modal from '@/components/UI/modal/BaseModal/Modal.vue'
import Button from '@/components/UI/buttons/Button/Button.vue'
import TextInput from '@/components/UI/inputs/TextInput/TextInput.vue'
import { useAuthStore } from '@/stores/auth'
import style from './NewChatModal.module.scss'
export default defineComponent({
  name: 'NewChatModal',
  components: { Modal, Button, TextInput },
  props: { modelValue: Boolean },
  emits: ['update:modelValue', 'select'],
  data() {
    return { style, auth: useAuthStore(), userId: '', opening: false, error: '' }
  },
  watch: {
    modelValue(open: boolean) {
      if (open) {
        this.userId = ''
        this.error = ''
      }
    },
  },
  methods: {
    async openDirect() {
      if (this.opening || !Number.isSafeInteger(Number(this.userId)) || Number(this.userId) <= 0)
        return
      this.opening = true
      this.error = ''
      try {
        const id = await this.$messageStore.openDirect(Number(this.userId))
        this.$emit('select', id)
        this.$emit('update:modelValue', false)
      } catch (error) {
        this.error = error instanceof Error ? error.message : 'Не удалось открыть диалог.'
      } finally {
        this.opening = false
      }
    },
    async openGroup() {
      if (this.opening || !this.auth.user?.student) return
      this.opening = true
      this.error = ''
      try {
        const id = await this.$messageStore.openGroup(this.auth.user.student.group_id)
        this.$emit('select', id)
        this.$emit('update:modelValue', false)
      } catch (error) {
        this.error = error instanceof Error ? error.message : 'Не удалось открыть чат группы.'
      } finally {
        this.opening = false
      }
    },
  },
})
</script>
