<script setup>
import { nextTick, ref, watch } from 'vue'

const props = defineProps({
  open: {
    type: Boolean,
    default: false,
  },
  title: {
    type: String,
    default: '确认操作',
  },
  message: {
    type: String,
    default: '',
  },
  confirmLabel: {
    type: String,
    default: '确认',
  },
  cancelLabel: {
    type: String,
    default: '取消',
  },
})

const emit = defineEmits(['confirm', 'cancel'])
const shellRef = ref(null)
const cancelRef = ref(null)
const confirmRef = ref(null)

async function focusPrimaryAction() {
  await nextTick()
  cancelRef.value?.focus()
}

function onKeydown(event) {
  if (event.key === 'Escape') {
    event.preventDefault()
    emit('cancel')
    return
  }
  if (event.key !== 'Tab') return

  const nodes = [cancelRef.value, confirmRef.value].filter(Boolean)
  if (nodes.length === 0) return
  const currentIndex = nodes.indexOf(document.activeElement)
  if (currentIndex === -1) {
    event.preventDefault()
    nodes[0].focus()
    return
  }
  event.preventDefault()
  const direction = event.shiftKey ? -1 : 1
  const nextIndex = (currentIndex + direction + nodes.length) % nodes.length
  nodes[nextIndex].focus()
}

watch(() => props.open, (isOpen) => {
  if (isOpen) {
    focusPrimaryAction()
  }
})
</script>

<template>
  <div v-if="open" class="modal-backdrop" @click.self="emit('cancel')">
    <section ref="shellRef" class="modal-card confirm-modal" tabindex="-1" @keydown="onKeydown">
      <div class="modal-header">
        <div>
          <h2>{{ title }}</h2>
          <p v-if="message">{{ message }}</p>
        </div>
      </div>
      <div class="modal-footer">
        <button ref="cancelRef" class="secondary-button" type="button" @click="emit('cancel')">{{ cancelLabel }}</button>
        <button ref="confirmRef" class="danger-button" type="button" @click="emit('confirm')">{{ confirmLabel }}</button>
      </div>
    </section>
  </div>
</template>
