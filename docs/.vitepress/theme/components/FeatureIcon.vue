<template>
  <component :is="icon" :size="size" :stroke-width="strokeWidth" />
</template>

<script setup>
import { Target, RefreshCw, Layers, Search, GitBranch, Shield } from 'lucide-vue-next'

const props = defineProps({
  name: String,
  size: {
    type: Number,
    default: 24
  },
  strokeWidth: {
    type: Number,
    default: 2
  }
})

const icons = {
  Target,
  RefreshCw,
  Layers,
  Search,
  GitBranch,
  Shield
}

const icon = icons[props.name]
</script>
