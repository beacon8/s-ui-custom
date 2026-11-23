// Plugins
import antd from './antd'

// Types
import type { App } from 'vue'

export function registerPlugins(app: App) {
  app.use(antd)
}
