<template>
  <section :class="style.chatBlock" aria-labelledby="dashboard-chat-title">
    <header :class="style.header">
      <h2 id="dashboard-chat-title" :class="style.heading">
        <PhChatDots :size="22" aria-hidden="true" />Чат
      </h2>
      <RouterLink v-press :class="style.allLink" to="/messages">
        Все чаты <PhArrowRight :size="14" aria-hidden="true" />
      </RouterLink>
    </header>
    <ul :class="style.list">
      <li v-for="chat in chats" :key="chat.id">
        <RouterLink
          v-press
          :class="style.chatItem"
          :to="{ path: '/messages', query: { chat: chat.id } }"
        >
          <span :class="[style.avatar, style[chat.avatar]]" aria-hidden="true">
            <PhUsersThree v-if="chat.group" :size="23" weight="fill" />
            <template v-else>{{ chat.initials }}</template>
            <span v-if="chat.online" :class="style.online" />
          </span>
          <span :class="style.copy">
            <span :class="style.name">{{ chat.name }}</span>
            <span :class="style.preview">{{ chat.preview }}</span>
          </span>
          <span :class="style.meta">
            <span :class="style.time">{{ chat.time }}</span>
            <span
              v-if="chat.unread"
              :class="style.unread"
              :aria-label="`${chat.unread} непрочитанных сообщения`"
            >
              {{ chat.unread }}
            </span>
          </span>
        </RouterLink>
      </li>
    </ul>
    <RouterLink v-press :class="style.compose" :to="{ path: '/messages', query: { compose: '1' } }">
      <PhChatDots :size="20" aria-hidden="true" />Написать сообщение
    </RouterLink>
  </section>
</template>

<script lang="ts">
import { defineComponent } from 'vue'
import { RouterLink } from 'vue-router'
import { PhChatDots, PhArrowRight, PhUsersThree } from '@phosphor-icons/vue'
import { pressFeedback } from '@/directives/pressFeedback'
import style from './ChatComponent.module.scss'

export default defineComponent({
  name: 'ChatComponent',
  components: { RouterLink, PhChatDots, PhArrowRight, PhUsersThree },
  directives: { press: pressFeedback },
  data() {
    return {
      style,
      chats: [
        {
          id: 'group-it-202',
          name: 'Группа_ИТ-202',
          preview: 'Иван: Кто будет на практике сегодня?',
          time: '10:24',
          avatar: 'slate',
          group: true,
          online: true,
          unread: 2,
        },
        {
          id: 'curator',
          name: 'Куратор группы',
          preview: 'Хорошо, спасибо!',
          time: '09:42',
          avatar: 'orange',
          initials: 'КГ',
        },
        {
          id: 'general',
          name: 'Общий чат',
          preview: 'Мария: Привет! Есть вопросы по заданию?',
          time: 'Вчера',
          avatar: 'blue',
          group: true,
        },
        {
          id: 'alexey',
          name: 'Алексей Смирнов',
          preview: 'Договорились, созвонимся завтра.',
          time: 'Вчера',
          avatar: 'slate',
          initials: 'АС',
        },
      ],
    }
  },
})
</script>
