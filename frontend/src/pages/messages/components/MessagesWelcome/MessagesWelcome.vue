<template>
  <div :class="style.welcome">
    <span :class="style.welcomeIcon">
      <PhChatsCircle :size="36" weight="light" aria-hidden="true" />
    </span>
    <h2>Хороший разговор начинается здесь</h2>
    <p>Выберите чат слева или напишите<br />кому-нибудь первым.</p>
    <Button variant="secondary" @click="$emit('compose')">
      <PhNotePencil :size="18" aria-hidden="true" />
      Новый диалог
    </Button>
  </div>
</template>

<script lang="ts">
import { defineComponent } from 'vue'
import { PhChatsCircle, PhNotePencil } from '@phosphor-icons/vue'
import Button from '@/components/UI/buttons/Button/Button.vue'
import style from './MessagesWelcome.module.scss'
export default defineComponent({
  name: 'MessagesWelcome',
  components: { PhChatsCircle, PhNotePencil, Button },
  emits: ['compose'],
  data() {
    return { style }
  },
})
</script>
