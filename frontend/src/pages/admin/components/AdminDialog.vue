<template>
  <Modal
    :model-value="modelValue"
    :width="width"
    :close-on-overlay="!$adminStore.saving"
    :swipe-to-close="!$adminStore.saving"
    @update:model-value="!$adminStore.saving && $emit('update:modelValue', $event)"
  >
    <section :class="style.dialog">
      <header :class="style.dialogHeader">
        <h2>{{ title }}</h2>
        <IconButton
          round
          title="Закрыть"
          variant="ghost"
          :disabled="$adminStore.saving"
          @click="$emit('update:modelValue', false)"
        >
          <PhX :size="20" />
        </IconButton>
      </header>
      <p v-if="$adminStore.error" :class="style.error" role="alert">{{ $adminStore.error }}</p>
      <slot />
    </section>
  </Modal>
</template>
<script lang="ts">
import { defineComponent } from 'vue'
import { PhX } from '@phosphor-icons/vue'
import Modal from '@/components/UI/modal/BaseModal/Modal.vue'
import IconButton from '@/components/UI/buttons/IconButton/IconButton.vue'
import style from '../Admin.module.scss'
export default defineComponent({
  components: { Modal, IconButton, PhX },
  props: {
    modelValue: Boolean,
    title: { type: String, required: true },
    width: { type: Number, default: 640 },
  },
  emits: ['update:modelValue'],
  data() {
    return { style }
  },
  watch: {
    modelValue(value) {
      if (value) this.$adminStore.error = ''
    },
  },
})
</script>
