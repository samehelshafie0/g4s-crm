<script setup lang="ts">
import { useRoute } from 'vue-router'
import MainLayout from '@/layouts/MainLayout.vue'
const route = useRoute()
</script>

<template>
  <RouterView v-if="route.meta.public" />
  <MainLayout v-else />
</template>
