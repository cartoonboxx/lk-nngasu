<template>
  <section :class="style.notFoundPage">
    <div :class="style.content">
      <div :class="style.textBlock">
        <h1>404</h1>
        <span>Страница не найдена</span>
      </div>
      <RouterLink to="dashboard">
        <!--        <Button variant="ghost"><PhHouse size="20" /> На главную</Button>-->
        <Tooltip text="На главную" position="bottom">
          <IconButton><PhHouse size="20" /></IconButton>
        </Tooltip>
      </RouterLink>
    </div>
  </section>
</template>

<script lang="ts">
import { defineComponent } from 'vue'
import style from './NotFoundPage.module.scss'
import { PhHouse } from '@phosphor-icons/vue'
import IconButton from '@/components/UI/buttons/IconButton/IconButton.vue'
import Tooltip from '@/components/UI/Tooltip/Tooltip.vue'

export default defineComponent({
  name: 'NotFoundPage',
  components: { Tooltip, IconButton, PhHouse },
  data() {
    return {
      style,
    }
  },
})
</script>
