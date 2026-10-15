<script setup lang="ts">
import { useRoute } from 'vue-router'
import MainLayout from '@/layouts/MainLayout.vue'
import EnvironmentBanner from '@/components/EnvironmentBanner.vue'
const route = useRoute()
</script>

<template>
  <RouterView v-if="route.meta.public" />
  <MainLayout v-else />
  <EnvironmentBanner />
</template>
