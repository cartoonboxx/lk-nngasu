import { mount, enableAutoUnmount } from '@vue/test-utils'
import { afterEach, describe, expect, it } from 'vitest'
import AssignmentModal from '@/pages/subjects/course/components/AssignmentModal/AssignmentModal.vue'
import type { AssignmentAnswer } from '@/pages/subjects/course/components/AssignmentModal/assignment.types'

const topic = {
  id: 'intro',
  title: 'База знаний',
  description: 'Описание',
  task: 'Создайте базу знаний',
  deadline: '15 октября',
  steps: ['Добавьте факты', 'Проверьте запросы'],
}
const answer: AssignmentAnswer = {
  attachments: [],
  comment: 'Мой ответ',
  submittedAt: '2026-10-19T10:00:00Z',
}
function render(value: AssignmentAnswer | null = null) {
  return mount(AssignmentModal, {
    props: { modelValue: true, topic, topicIndex: 0, answer: value },
    global: { stubs: { Modal: { template: '<div><slot /></div>' } } },
  })
}
enableAutoUnmount(afterEach)

describe('AssignmentModal', () => {
  it('requires a comment or attachment and emits the answer without mutating props', async () => {
    const wrapper = render()
    expect(wrapper.get('button[type="submit"]').attributes('disabled')).toBeDefined()
    await wrapper.get('textarea').setValue('  Готовая работа  ')
    await wrapper.get('form').trigger('submit')
    expect(wrapper.emitted('submit')).toEqual([[{ attachments: [], comment: 'Готовая работа' }]])
  })

  it('adds several files, ignores duplicates and removes individual attachments', async () => {
    const wrapper = render()
    const input = wrapper.get('input[type="file"]')
    const file = new File(['facts'], 'answer.pl', { lastModified: 1 })
    const report = new File(['report'], 'report.txt', { lastModified: 2 })
    Object.defineProperty(input.element, 'files', {
      value: [file, file, report],
      configurable: true,
    })
    await input.trigger('change')
    expect(wrapper.findAll('li').filter((item) => item.text().includes('answer.pl'))).toHaveLength(
      1,
    )
    await wrapper.get('[aria-label="Удалить файл report.txt"]').trigger('click')
    await wrapper.get('form').trigger('submit')
    const submitted = wrapper.emitted('submit')![0]![0] as AssignmentAnswer
    expect(submitted.attachments).toHaveLength(1)
    expect(submitted.attachments[0]!.file).toBe(file)
  })

  it('rejects oversized files', async () => {
    const wrapper = render()
    const file = new File(['data'], 'large.zip')
    Object.defineProperty(file, 'size', { value: 21 * 1024 * 1024 })
    const input = wrapper.get('input[type="file"]')
    Object.defineProperty(input.element, 'files', { value: [file] })
    await input.trigger('change')
    expect(wrapper.get('[role="alert"]').text()).toContain('превышает 20 МБ')
    expect(wrapper.get('button[type="submit"]').attributes('disabled')).toBeDefined()
  })

  it('cancels edits without changing the submitted answer and confirms deletion', async () => {
    const wrapper = render(answer)
    const button = (text: string) => wrapper.findAll('button').find((item) => item.text() === text)!
    await button('Редактировать ответ').trigger('click')
    await wrapper.get('textarea').setValue('Новый ответ')
    await button('Отменить').trigger('click')
    expect(wrapper.text()).toContain('Мой ответ')
    expect(wrapper.emitted('submit')).toBeUndefined()
    expect(answer.comment).toBe('Мой ответ')
    await button('Удалить ответ').trigger('click')
    expect(wrapper.emitted('delete')).toBeUndefined()
    await button('Подтвердить удаление').trigger('click')
    expect(wrapper.emitted('delete')).toHaveLength(1)
    await wrapper.setProps({ answer: null })
    expect(wrapper.find('form').exists()).toBe(true)
    expect(wrapper.get('textarea').element.value).toBe('')
  })

  it('shows the grade and teacher feedback when a submitted answer is reviewed', async () => {
    const wrapper = render(answer)
    await wrapper
      .findAll('button')
      .find((item) => item.text() === 'Редактировать ответ')!
      .trigger('click')
    await wrapper.setProps({
      answer: { ...answer, review: { grade: 0, comment: 'Дополните правила' } },
    })
    expect(wrapper.get('[aria-label="Результат проверки"]').text()).toContain('0')
    expect(wrapper.text()).toContain('Дополните правила')
    expect(wrapper.text()).not.toContain('Редактировать ответ')
    expect(wrapper.text()).not.toContain('Удалить ответ')
    expect(wrapper.find('form').exists()).toBe(false)
  })
})
