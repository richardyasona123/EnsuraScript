import DefaultTheme from 'vitepress/theme'
import './style.css'
import { Target, RefreshCw, Layers, Search, GitBranch, Shield } from 'lucide-vue-next'

export default {
  extends: DefaultTheme,
  enhanceApp({ app }) {
    app.component('Target', Target)
    app.component('RefreshCw', RefreshCw)
    app.component('Layers', Layers)
    app.component('Search', Search)
    app.component('GitBranch', GitBranch)
    app.component('Shield', Shield)
  }
}
