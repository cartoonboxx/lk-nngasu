<template>
  <Modal
    :model-value="modelValue"
    @update:model-value="$emit('update:modelValue', $event)"
    :width="440"
  >
    <header :class="style.modalHeader">
      <div>
        <h2>Новый диалог</h2>
        <p>Кому хотите написать?</p>
      </div>
      <IconButton
        variant="ghost"
        size="sm"
        :class="style.headerAction"
        round
        @click="$emit('update:modelValue', false)"
      >
        <PhX :size="20" />
      </IconButton>
    </header>
    <Input v-model="contactSearch" placeholder="Имя или фамилия" full-width>
      <template #prefix>
        <PhMagnifyingGlass :size="18" aria-hidden="true" />
      </template>
    </Input>
    <ul :class="style.contacts">
      <li v-for="contact in filteredContacts" :key="contact.id">
        <StyleTab :class="style.contact" @select="$emit('select', contact.id)">
          <span :class="style.contactContent">
            <ChatAvatar :initials="contact.initials" />
            <span :class="style.contactCopy">
              <strong>{{ contact.name }}</strong>
              <span>{{ contact.description }}</span>
            </span>
          </span>
          <template #trailing>
            <PhArrowUpRight :size="17" aria-hidden="true" />
          </template>
        </StyleTab>
      </li>
    </ul>
    <p v-if="!filteredContacts.length" :class="style.contactsEmpty">
      Никого не нашли. Попробуйте другое имя.
    </p>
  </Modal>
</template>

<script lang="ts">
import { defineComponent } from 'vue'
import { PhX, PhMagnifyingGlass, PhArrowUpRight } from '@phosphor-icons/vue'
import Modal from '@/components/UI/modal/BaseModal/Modal.vue'
import IconButton from '@/components/UI/buttons/IconButton/IconButton.vue'
import Input from '@/components/UI/inputs/Input/Input.vue'
import StyleTab from '@/components/UI/tabs/StyleTab/StyleTab.vue'
import ChatAvatar from '../ChatAvatar/ChatAvatar.vue'
import { contacts } from '../../messages.data'
import style from './NewChatModal.module.scss'
export default defineComponent({
  name: 'NewChatModal',
  components: {
    PhX,
    PhMagnifyingGlass,
    PhArrowUpRight,
    Modal,
    IconButton,
    Input,
    StyleTab,
    ChatAvatar,
  },
  props: { modelValue: Boolean },
  emits: {
    'update:modelValue': (value: boolean) => typeof value === 'boolean',
    select: (id: string) => !!id,
  },
  data() {
    return { style, contactSearch: '' }
  },
  computed: {
    filteredContacts(): typeof contacts {
      return contacts.filter((contact) =>
        contact.name
          .toLocaleLowerCase('ru')
          .includes(this.contactSearch.trim().toLocaleLowerCase('ru')),
      )
    },
  },
  watch: {
    modelValue(open: boolean) {
      if (open) this.contactSearch = ''
    },
  },
})
</script>
