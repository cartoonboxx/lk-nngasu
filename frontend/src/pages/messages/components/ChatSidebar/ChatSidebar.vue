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
        <template #prefix><PhMagnifyingGlass :size="18" aria-hidden="true" /></template>
      </Input>
    </div>
    <div :class="style.filters" role="group" aria-label="Фильтр чатов">
      <Button
        size="sm"
        round
        variant="ghost"
        :aria-pressed="!unreadOnly"
        @click="unreadOnly = false"
      >
        Все
      </Button>
      <Button size="sm" round variant="ghost" :aria-pressed="unreadOnly" @click="unreadOnly = true">
        Непрочитанные
        <Badge v-if="unreadChats" :value="unreadChats" inline />
      </Button>
    </div>
    <ul
      v-if="filteredChats.length"
      ref="chatList"
      :class="style.chatList"
      aria-label="Список диалогов"
    >
      <li v-for="chat in filteredChats" :key="chat.id">
        <StyleTab
          :class="style.chatRow"
          :active="selectedId === chat.id"
          :trigger-attrs="{
            'aria-current': selectedId === chat.id ? 'true' : undefined,
            'data-chat-id': chat.id,
          }"
          @select="$emit('select', chat.id)"
        >
          <span :class="style.chatContent">
            <ChatAvatar :initials="chat.initials" :group="chat.group" />
            <span :class="style.chatCopy">
              <span :class="style.chatLine">
                <span :class="[style.chatName, chat.unread && style.unreadName]">
                  {{ chat.name }}
                </span>
                <time v-if="lastMessage(chat)" :class="style.chatTime">
                  {{ listTime(lastMessage(chat)!.sentAt) }}
                </time>
              </span>
              <span :class="style.chatLine">
                <span :class="style.preview">{{ preview(chat) }}</span>
                <Badge
                  v-if="chat.unread"
                  :value="chat.unread"
                  :aria-label="`Непрочитанных сообщений: ${chat.unread}`"
                  inline
                />
              </span>
            </span>
          </span>
        </StyleTab>
      </li>
    </ul>
    <div v-else :class="style.listEmpty">
      <PhMagnifyingGlass :size="25" aria-hidden="true" />
      <p>{{ search ? 'Чаты не найдены' : 'Всё прочитано' }}</p>
      <span>
        {{ search ? 'Попробуйте другое имя или название' : 'Новые сообщения появятся здесь' }}
      </span>
    </div>
    <p :class="style.localNote">Переписки сохраняются в этом браузере</p>
  </aside>
</template>

<script lang="ts">
import { defineComponent, type PropType } from 'vue'
import style from './ChatSidebar.module.scss'
import { PhNotePencil, PhMagnifyingGlass } from '@phosphor-icons/vue'
import IconButton from '@/components/UI/buttons/IconButton/IconButton.vue'
import Button from '@/components/UI/buttons/Button/Button.vue'
import Input from '@/components/UI/inputs/Input/Input.vue'
import StyleTab from '@/components/UI/tabs/StyleTab/StyleTab.vue'
import Badge from '@/components/UI/Badge/Badge.vue'
import ChatAvatar from '../ChatAvatar/ChatAvatar.vue'
import type { Chat } from '../../messages.types'
import { lastMessage, messageTime, dayLabel } from '../../messages.format'
export default defineComponent({
  name: 'ChatSidebar',
  components: {
    PhNotePencil,
    PhMagnifyingGlass,
    IconButton,
    Button,
    Input,
    StyleTab,
    Badge,
    ChatAvatar,
  },
  props: { chats: { type: Array as PropType<Chat[]>, required: true }, selectedId: String },
  emits: { select: (id: string) => !!id, compose: () => true },
  data() {
    return { style, search: '', unreadOnly: false }
  },
  computed: {
    unreadChats(): number {
      return this.chats.filter((chat) => chat.unread > 0).length
    },
    filteredChats(): Chat[] {
      return this.chats
        .filter(
          (chat) =>
            (!this.unreadOnly || chat.unread > 0) &&
            chat.name.toLocaleLowerCase('ru').includes(this.search.trim().toLocaleLowerCase('ru')),
        )
        .sort((a, b) => this.chatTimestamp(b) - this.chatTimestamp(a))
    },
  },
  methods: {
    lastMessage,
    chatTimestamp(chat: Chat): number {
      const message = lastMessage(chat)
      return message ? Date.parse(message.sentAt) : Number.MAX_SAFE_INTEGER
    },
    preview(chat: Chat): string {
      const message = lastMessage(chat)
      if (!message) return 'Начните разговор'
      return `${message.own ? 'Вы: ' : chat.group ? `${message.sender.split(' ')[0]}: ` : ''}${message.text}`
    },
    listTime(date: string): string {
      return new Date(date).toDateString() === new Date().toDateString()
        ? messageTime(date)
        : dayLabel(date)
    },
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
