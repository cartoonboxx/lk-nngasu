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
    }
  },
  computed: {
    chats() {
      return this.$messageStore.chats
        .map((chat) => {
          const last = chat.messages[chat.messages.length - 1]
          return {
            ...chat,
            sender: last?.own ? 'Вы' : chat.group ? last?.sender.split(' ')[0] : undefined,
            preview: last?.text ?? 'Начните разговор',
            time: last
              ? new Date(last.sentAt).toDateString() === new Date().toDateString()
                ? new Date(last.sentAt).toLocaleTimeString('ru', {
                    hour: '2-digit',
                    minute: '2-digit',
                  })
                : new Date(last.sentAt).toLocaleDateString('ru', { day: 'numeric', month: 'short' })
              : '',
            timestamp: last ? Date.parse(last.sentAt) : 0,
          }
        })
        .sort((a, b) => b.timestamp - a.timestamp)
        .slice(0, 4)
    },
  },
})
</script>
