<script setup lang="ts">
import type { Plan } from '../api/billing'

const props = defineProps<{
  plan: Plan
  isCurrent: boolean
  isPopular?: boolean
}>()

const emit = defineEmits<{
  select: [plan: Plan]
}>()

function buttonLabel(): string {
  if (props.isCurrent) return 'Current Plan'
  return 'Select Plan'
}

function featureLabel(plan: Plan): string[] {
  const items: string[] = []
  items.push(plan.appLimit === -1 ? 'Unlimited apps' : `${plan.appLimit} apps`)
  items.push(plan.dbLimit === -1 ? 'Unlimited databases' : `${plan.dbLimit} databases`)
  items.push(`${plan.memoryMb} MB memory`)
  items.push(`${plan.cpuCores} CPU cores`)
  items.push(`${plan.storageGb} GB storage`)
  if (plan.customDomains) items.push('Custom domains')
  if (plan.prioritySupport) items.push('Priority support')
  return items
}

const features = featureLabel(props.plan)
const isBusiness = props.plan.name === 'business'
</script>

<template>
  <div
    class="relative flex flex-col rounded-2xl border-2 p-6 transition-all"
    :class="[
      isCurrent
        ? 'border-indigo-500 shadow-lg shadow-indigo-100'
        : isBusiness
          ? 'border-gray-800 bg-gray-900 shadow-lg'
          : 'border-gray-200 bg-white shadow hover:shadow-md',
    ]"
  >
    <!-- Popular badge -->
    <div
      v-if="isPopular"
      class="absolute -top-3 left-1/2 -translate-x-1/2 rounded-full bg-indigo-600 px-4 py-1 text-xs font-semibold text-white"
    >
      Popular
    </div>

    <!-- Current badge -->
    <div
      v-if="isCurrent"
      class="absolute -top-3 right-4 rounded-full bg-indigo-100 px-3 py-1 text-xs font-semibold text-indigo-700"
    >
      Current
    </div>

    <!-- Plan name -->
    <h3
      class="text-lg font-semibold"
      :class="isBusiness && !isCurrent ? 'text-white' : 'text-gray-900'"
    >
      {{ plan.displayName }}
    </h3>

    <!-- Price -->
    <div class="mt-4 flex items-baseline gap-1">
      <span
        class="text-4xl font-bold tracking-tight"
        :class="isBusiness && !isCurrent ? 'text-white' : 'text-gray-900'"
      >
        &euro;{{ plan.priceEur }}
      </span>
      <span
        class="text-sm font-medium"
        :class="isBusiness && !isCurrent ? 'text-gray-400' : 'text-gray-500'"
      >
        /mo
      </span>
    </div>

    <!-- Features -->
    <ul class="mt-6 flex-1 space-y-3">
      <li
        v-for="(feat, i) in features"
        :key="i"
        class="flex items-start gap-2 text-sm"
        :class="isBusiness && !isCurrent ? 'text-gray-300' : 'text-gray-600'"
      >
        <svg
          class="mt-0.5 h-4 w-4 shrink-0"
          :class="isBusiness && !isCurrent ? 'text-indigo-400' : 'text-indigo-500'"
          fill="none"
          stroke="currentColor"
          viewBox="0 0 24 24"
        >
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
        </svg>
        {{ feat }}
      </li>
    </ul>

    <!-- Action button -->
    <button
      class="mt-6 w-full rounded-lg px-4 py-2.5 text-sm font-semibold transition-colors"
      :class="
        isCurrent
          ? 'cursor-default bg-indigo-50 text-indigo-400'
          : isPopular
            ? 'bg-indigo-600 text-white hover:bg-indigo-700'
            : isBusiness && !isCurrent
              ? 'bg-white text-gray-900 hover:bg-gray-100'
              : 'bg-indigo-50 text-indigo-600 hover:bg-indigo-100'
      "
      :disabled="isCurrent"
      @click="emit('select', plan)"
    >
      {{ isCurrent ? 'Current Plan' : buttonLabel() }}
    </button>
  </div>
</template>
