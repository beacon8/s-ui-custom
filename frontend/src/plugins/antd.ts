import Antd from 'ant-design-vue'
import 'ant-design-vue/dist/reset.css'
import AppIcon from '@/components/AppIcon.vue'
import type { App } from 'vue'

export default {
  install(app: App) {
    app.use(Antd)
    app.component('AppIcon', AppIcon)
  },
}
