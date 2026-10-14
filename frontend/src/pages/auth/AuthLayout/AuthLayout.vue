<template>
  <main :class="style.page">
    <div :class="style.shell">
      <aside :class="style.hero" aria-label="Личный кабинет ННГАСУ">
        <div :class="style.brand">
          <span :class="style.brandMark"><img src="$images/logo.svg" alt="" /></span>
          <strong :class="style.universityName">ННГАСУ</strong>
          <div :class="style.portalName">
            <strong>Личный кабинет</strong>
            <span>студента и сотрудника</span>
          </div>
        </div>

        <div :class="style.heroContent">
          <h1 :class="style.headline">Твой университет.<br />Всегда <span>рядом.</span></h1>
          <p :class="style.description">
            Расписание, учебные материалы и важные новости — в одном пространстве.
          </p>
        </div>

        <div :class="style.illustration" aria-hidden="true">
          <span :class="style.orbit" />
          <span :class="style.orbitInner" />
          <svg :class="style.architecture" viewBox="0 0 560 290" fill="none">
            <g :class="style.drawingGuides" stroke="currentColor" stroke-width="0.8">
              <path
                d="M12 234 284 76l265 152M64 265 330 111M114 286 381 133M6 181l263 152M85 134l264 152"
              />
              <path d="M279 20v250M113 77v174M450 75v181" stroke-dasharray="3 7" />
              <circle cx="279" cy="172" r="139" stroke-dasharray="3 7" />
            </g>
            <path :class="style.buildingShadow" d="m85 226 192-110 224 105-194 68z" />
            <g stroke-linejoin="round" stroke-width="1.5">
              <path :class="style.buildingFront" d="m98 189 84 46v25l-84-46z" />
              <path :class="style.buildingSide" d="m182 235 47-26v25l-47 26z" />
              <path :class="style.buildingRoof" d="m98 189 47-26 84 46-47 26z" />
              <path :class="style.buildingFront" d="m405 164 50 27v28l-50-27z" />
              <path :class="style.buildingSide" d="m455 191 47-26v28l-47 26z" />
              <path :class="style.buildingRoof" d="m405 164 47-26 50 27-47 26z" />
              <path :class="style.buildingFront" d="m143 143 129 71v56l-129-71z" />
              <path :class="style.buildingSide" d="m272 214 133-76v56l-133 76z" />
              <path :class="style.buildingRoof" d="m143 143 133-76 129 71-133 76z" />
              <path :class="style.buildingFront" d="m180 104 91 50v37l-91-50z" />
              <path :class="style.buildingSide" d="m271 154 96-54v37l-96 54z" />
              <path :class="style.buildingRoof" d="m180 104 96-54 91 50-96 54z" />
            </g>
            <g
              :class="style.windows"
              stroke="currentColor"
              stroke-width="1.3"
              stroke-linejoin="round"
            >
              <path
                d="m160 166 11 6v15l-11-6zm26 14 11 6v15l-11-6zm26 15 11 6v15l-11-6zm26 14 11 6v15l-11-6z"
              />
              <path
                d="m160 189 11 6v9l-11-6zm26 14 11 6v9l-11-6zm26 15 11 6v9l-11-6zm26 14 11 6v9l-11-6z"
              />
              <path
                d="m289 218 12-7v15l-12 7zm30-17 12-7v15l-12 7zm30-17 12-7v15l-12 7zm30-17 12-7v15l-12 7z"
              />
              <path
                d="m289 241 12-7v9l-12 7zm30-17 12-7v9l-12 7zm30-17 12-7v9l-12 7zm30-17 12-7v9l-12 7z"
              />
              <path d="m196 124 10 5v13l-10-5zm26 14 10 5v13l-10-5zm26 14 10 5v13l-10-5z" />
              <path d="m286 157 10-6v13l-10 6zm26-15 10-6v13l-10 6zm26-15 10-6v13l-10 6z" />
              <path d="m111 206 10 5m5-2 10 5m5-2 10 5m314-15 12-7m4 5 12-7" />
            </g>
            <g :class="style.drawingAccent" stroke="currentColor" stroke-width="2">
              <path d="m143 186 129 71 133-76M180 139l91 50 96-54" />
              <path d="M276 50V18m0 0 28 10-28 8" />
              <circle cx="455" cy="219" r="4" fill="currentColor" stroke="none" />
              <path d="M467 244h35m-17-17v35" stroke-width="1" />
            </g>
          </svg>
          <div :class="style.illustrationLabel"><span /> Пространство твоих возможностей</div>
        </div>

        <div :class="style.services">
          <span><PhCalendarBlank :size="17" aria-hidden="true" /> Расписание</span>
          <span><PhBookOpen :size="17" aria-hidden="true" /> Учёба</span>
          <span><PhBell :size="17" aria-hidden="true" /> Новости</span>
        </div>
        <p :class="style.heroFooter">Нижний Новгород <span>·</span> ННГАСУ</p>
      </aside>

      <section :class="style.panel" :aria-labelledby="titleId">
        <div :class="style.panelTop">
          <PhStudent :size="18" aria-hidden="true" /> Твой цифровой кампус
        </div>
        <div :class="style.formArea">
          <slot />
        </div>
        <p :class="style.panelFooter">
          Личный кабинет ННГАСУ <span>Учись. Создавай. Двигайся вперёд.</span>
        </p>
      </section>
    </div>
  </main>
</template>

<script lang="ts">
import { PhBell, PhBookOpen, PhCalendarBlank, PhStudent } from '@phosphor-icons/vue'
import style from './AuthLayout.module.scss'
import { defineComponent } from 'vue'

export default defineComponent({
  name: 'AuthLayout',
  components: { PhBell, PhBookOpen, PhCalendarBlank, PhStudent },
  props: {
    titleId: {
      type: String,
      required: true,
    },
  },
  data() {
    return {
      style,
    }
  },
})
</script>
