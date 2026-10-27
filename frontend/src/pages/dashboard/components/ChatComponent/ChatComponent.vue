<template>
  <section :class="style.chatBlock" aria-labelledby="dashboard-chat-title">
    <header :class="style.header">
      <h2 id="dashboard-chat-title" :class="style.heading">Чаты</h2>
      <RouterLink v-press :class="style.allLink" to="/messages">
        Все <PhArrowUpRight :size="15" aria-hidden="true" />
      </RouterLink>
    </header>
    <ul v-if="chats.length" :class="style.list" aria-label="Последние диалоги">
      <li v-for="chat in chats" :key="chat.id">
        <RouterLink
          v-press
          :class="[style.chatItem, chat.unread && style.hasUnread]"
          :to="{ path: '/messages', query: { chat: chat.id } }"
        >
          <span :class="style.avatar" aria-hidden="true">
            <PhUsersThree v-if="chat.group" :size="21" />
            <template v-else>{{ chat.initials }}</template>
          </span>
          <span :class="style.copy">
            <span :class="style.topLine">
              <span :class="style.name">{{ chat.name }}</span>
              <span :class="style.time">{{ chat.time }}</span>
            </span>
            <span :class="style.bottomLine">
              <span :class="style.preview">
                <span v-if="chat.sender" :class="style.sender">{{ chat.sender }}: </span>
                {{ chat.preview }}
              </span>
              <span
                v-if="chat.unread"
                :class="style.unread"
                :aria-label="`Непрочитанных сообщений: ${chat.unread}`"
              >
                {{ chat.unread }}
              </span>
            </span>
          </span>
        </RouterLink>
      </li>
    </ul>
    <div v-else :class="style.empty">
      <PhChatsCircle :size="32" aria-hidden="true" />
      <p>Пока нет диалогов</p>
      <span>Напишите одногруппнику или куратору</span>
    </div>
    <footer :class="style.footer">
      <RouterLink
        v-press
        :class="style.compose"
        :to="{ path: '/messages', query: { compose: '1' } }"
      >
        <PhNotePencil :size="18" aria-hidden="true" />
        Написать сообщение
      </RouterLink>
    </footer>
  </section>
</template>

<script lang="ts">
import { defineComponent } from 'vue'
import { RouterLink } from 'vue-router'
import { PhChatsCircle, PhNotePencil, PhArrowUpRight, PhUsersThree } from '@phosphor-icons/vue'
import { pressFeedback } from '@/directives/pressFeedback'
import style from './ChatComponent.module.scss'

export default defineComponent({
  name: 'ChatComponent',
  components: { RouterLink, PhChatsCircle, PhNotePencil, PhArrowUpRight, PhUsersThree },
  directives: { press: pressFeedback },
  data() {
    return {
      style,
      chats: [
        {
          id: 'group-it-202',
          name: 'Группа ИТ-202',
          sender: 'Иван',
          preview: 'Кто будет на практике сегодня?',
          time: '10:24',
          group: true,
          unread: 2,
        },
        {
          id: 'curator',
          name: 'Куратор группы',
          preview: 'Хорошо, спасибо!',
          time: '09:42',
          initials: 'КГ',
        },
        {
          id: 'general',
          name: 'Общий чат',
          sender: 'Мария',
          preview: 'Привет! Есть вопросы по заданию?',
          time: 'Вчера',
          group: true,
        },
        {
          id: 'alexey',
          name: 'Алексей Смирнов',
          preview: 'Договорились, созвонимся завтра.',
          time: 'Вчера',
          initials: 'АС',
        },
      ],
    }
  },
})
</script>
