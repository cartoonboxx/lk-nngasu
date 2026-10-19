<template>
  <section :class="style.coursePage">
    <SectionHeader
      header="Логическое программирование"
      paragraph="7 семестр · Учебные материалы курса"
    />
    <div :class="style.courseContent">
      <AsideCoursePanel
        :topics="topics"
        :active-topic-index="activeTopicIndex"
        @select="selectTopic"
      />
      <ContentCourseComponent
        :topic="activeTopic"
        :topic-index="activeTopicIndex"
        :topic-count="topics.length"
        @open-assignment="assignmentOpen = true"
        @previous="selectTopic(activeTopicIndex - 1)"
        @next="selectTopic(activeTopicIndex + 1)"
      />
    </div>
    <AssignmentModal
      v-model="assignmentOpen"
      course-title="Логическое программирование"
      :topic="activeTopic"
      :topic-index="activeTopicIndex"
      :answer="activeAnswer"
      @submit="submitAnswer"
      @delete="deleteAnswer"
    />
  </section>
</template>

<script lang="ts">
import { defineComponent } from 'vue'
import SectionHeader from '@/components/SectionHeader/SectionHeader.vue'
import AsideCoursePanel from './components/AsideCoursePanel/AsideCoursePanel.vue'
import ContentCourseComponent from './components/ContentCourseComponent/ContentCourseComponent.vue'
import AssignmentModal from './components/AssignmentModal/AssignmentModal.vue'
import type {
  AssignmentAnswer,
  AssignmentAnswerDraft,
} from './components/AssignmentModal/assignment.types'
import { assignmentAnswers } from './assignmentAnswers'
import type { CourseTopic } from './course.types'
import style from './CoursePage.module.scss'

export default defineComponent({
  name: 'CoursePage',
  components: { SectionHeader, AsideCoursePanel, ContentCourseComponent, AssignmentModal },
  data() {
    return {
      style,
      activeTopicIndex: 0,
      assignmentOpen: false,
      assignmentAnswers,
      topics: [
        {
          id: 'introduction',
          title: 'Знакомство с логическим программированием',
          description:
            'Изучите основные понятия логического программирования: факты, правила и запросы. Затем выполните практическую работу.',
          task: 'Создайте базу знаний и составьте запросы на языке Пролог.',
          deadline: '15 октября',
          steps: [
            'Опишите факты о членах семьи и их родственных связях.',
            'Добавьте правила для определения родства.',
            'Составьте не менее пяти запросов и проверьте результаты.',
          ],
        },
        {
          id: 'syntax',
          title: 'Синтаксис и значение программ языка Пролог',
          description:
            'Разберите структуру программы, термы, переменные и механизм унификации. Научитесь описывать отношения с помощью правил.',
          task: 'Опишите предметную область с помощью фактов и правил языка Пролог.',
          deadline: '22 октября',
          steps: [
            'Выберите предметную область и опишите её сущности.',
            'Запишите факты и правила, используя переменные и составные термы.',
            'Проверьте унификацию на примерах запросов.',
          ],
        },
        {
          id: 'lists',
          title: 'Списки и рекурсия',
          description:
            'Познакомьтесь со списками, их структурой и рекурсивной обработкой. Рассмотрите базовый случай и шаг рекурсии.',
          task: 'Реализуйте рекурсивные предикаты для обработки списков.',
          deadline: '29 октября',
          steps: [
            'Реализуйте определение длины списка.',
            'Добавьте поиск элемента и объединение двух списков.',
            'Проверьте предикаты на пустом списке и списках разной длины.',
          ],
        },
        {
          id: 'search',
          title: 'Поиск решений и отсечение',
          description:
            'Изучите поиск с возвратом и управление перебором решений. Разберите применение оператора отсечения.',
          task: 'Исследуйте поиск с возвратом и оптимизируйте программу с помощью отсечения.',
          deadline: '5 ноября',
          steps: [
            'Проследите порядок поиска решений на выбранном примере.',
            'Добавьте отсечение и сравните результаты.',
            'Объясните, как изменились набор решений и порядок вычислений.',
          ],
        },
        {
          id: 'knowledge-base',
          title: 'Работа с базой знаний',
          description:
            'Объедините изученные приёмы и <i>создайте</i> базу знаний. Научитесь добавлять, изменять и запрашивать сведения.',
          task: 'Разработайте небольшую базу знаний для выбранной предметной области.',
          deadline: '12 ноября',
          steps: [
            'Спроектируйте структуру базы знаний и набор отношений.',
            'Реализуйте правила получения новых сведений.',
            'Подготовьте примеры использования и краткий отчёт.',
          ],
        },
      ] as CourseTopic[],
    }
  },
  computed: {
    answerKey(): string {
      return JSON.stringify([this.$route.params.id, this.activeTopic.id])
    },
    activeAnswer(): AssignmentAnswer | null {
      return this.assignmentAnswers[this.answerKey] ?? null
    },
    activeTopic(): CourseTopic {
      return this.topics[this.activeTopicIndex]!
    },
  },
  methods: {
    submitAnswer(draft: AssignmentAnswerDraft) {
      if (this.activeAnswer?.review) return
      this.assignmentAnswers[this.answerKey] = { ...draft, submittedAt: new Date().toISOString() }
    },
    deleteAnswer() {
      if (!this.activeAnswer?.review) delete this.assignmentAnswers[this.answerKey]
    },
    selectTopic(index: number) {
      if (index < 0 || index >= this.topics.length) return
      this.activeTopicIndex = index
    },
  },
})
</script>
