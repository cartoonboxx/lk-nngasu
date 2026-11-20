<template>
  <aside :class="style.sidebar">
    <header :class="style.sidebarHeader">
      <h2 ref="sidebarTitle" tabindex="-1">
        Чаты
        <span :class="style.chatCount">{{ chats.length }}</span>
      </h2>
      <IconButton
        variant="ghost"
        size="sm"
        round
        :class="style.headerAction"
        title="Новый диалог"
        @click="$emit('compose')"
      >
        <PhNotePencil :size="21" />
      </IconButton>
    </header>
    <div :class="style.search">
      <Input v-model="search" placeholder="Поиск по чатам" full-width>
        <template #prefix><PhMagnifyingGlass :size="18" /></template>
      </Input>
    </div>
    <ul v-if="filteredChats.length" ref="chatList" :class="style.chatList">
      <li v-for="chat in filteredChats" :key="chat.id">
        <div :class="style.chatWrapper">
          <StyleTab
            :class="[style.chatRow, chat.unreadCount > 0 && style.unreadChat]"
            :active="selectedId === chat.id"
            :to="{ name: 'messages', query: { chat: chat.id } }"
            :trigger-attrs="{
              'data-chat-id': chat.id,
            }"
          >
            <span :class="style.chatContent">
              <ChatAvatar :initials="chat.initials" :group="chat.group" />
              <span :class="style.chatCopy">
                <span :class="style.chatLine">
                  <span :class="[style.chatName, chat.unreadCount > 0 && style.unreadName]">
                    {{ chat.name }}
                  </span>
                  <time v-if="chat.lastMessage" :class="style.chatTime">
                    {{ chat.time }}
                  </time>
                </span>
                <span :class="style.chatLine">
                  <span :class="style.preview">{{ chat.preview }}</span>
                </span>
              </span>
            </span>
            <template #trailing>
              <Badge v-if="chat.unreadCount > 0" :value="chat.unreadCount" inline />
            </template>
          </StyleTab>
        </div>
      </li>
    </ul>
    <div v-else :class="style.listEmpty">
      <PhMagnifyingGlass :size="25" />
      <p>{{ search ? 'Чаты не найдены' : 'Пока нет диалогов' }}</p>
      <span>
        {{ search ? 'Попробуйте другое имя или название' : 'Новые сообщения появятся здесь' }}
      </span>
    </div>
    <p :class="style.localNote">Переписки сохраняются на сервере</p>
  </aside>
</template>

<script lang="ts">
import { defineComponent, type PropType } from 'vue'
import style from './ChatSidebar.module.scss'
import { PhNotePencil, PhMagnifyingGlass } from '@phosphor-icons/vue'
import IconButton from '@/components/UI/buttons/IconButton/IconButton.vue'
import Badge from '@/components/UI/Badge/Badge.vue'
import Input from '@/components/UI/inputs/Input/Input.vue'
import StyleTab from '@/components/UI/tabs/StyleTab/StyleTab.vue'
import ChatAvatar from '../ChatAvatar/ChatAvatar.vue'
import type { Chat } from '../../messages.types'
import { messageTimeFormatter, dayLabel } from '../../messages.format'
export default defineComponent({
  name: 'ChatSidebar',
  components: {
    PhNotePencil,
    PhMagnifyingGlass,
    IconButton,
    Badge,
    Input,
    StyleTab,
    ChatAvatar,
  },
  props: { chats: { type: Array as PropType<Chat[]>, required: true }, selectedId: String },
  emits: ['compose'],
  data() {
    return { style, search: '' }
  },
  computed: {
    filteredChats() {
      const query = this.search.trim().toLocaleLowerCase('ru')
      const today = new Date().toDateString()
      return this.chats
        .filter((chat) => chat.name.toLocaleLowerCase('ru').includes(query))
        .map((chat) => {
          const message = chat.lastMessage
          return {
            ...chat,
            lastMessage: message,
            timestamp: message ? Date.parse(message.sentAt) : 0,
            preview: message
              ? `${message.own ? 'Вы: ' : chat.group ? `${message.sender.split(' ')[0]}: ` : ''}${message.text}`
              : 'Начните разговор',
            time: !message
              ? ''
              : new Date(message.sentAt).toDateString() === today
                ? messageTimeFormatter.format(new Date(message.sentAt))
                : dayLabel(message.sentAt),
          }
        })
        .sort((a, b) => b.timestamp - a.timestamp)
    },
  },
  methods: {
    restoreFocus(id: string) {
      const list = this.$refs.chatList as HTMLElement | undefined
      const target = [...(list?.querySelectorAll<HTMLElement>('[data-chat-id]') ?? [])].find(
        (item) => item.dataset.chatId === id,
      )
      ;(target ?? (this.$refs.sidebarTitle as HTMLElement | undefined))?.focus({
        preventScroll: true,
      })
    },
  },
})
</script>
