<script setup lang="ts">
// About: what x-Ray VPN is, its version, where the docs and the discussions are.
import { useI18n } from "vue-i18n";
import {
  mdiBookOpenPageVariantOutline,
  mdiForumOutline,
  mdiGithub,
} from "@mdi/js";
import { useAppStore } from "@/stores/app";

defineOptions({ name: "AboutView" });
const { t } = useI18n();
const store = useAppStore();

const repo = "https://github.com/wywywywycloud/v2rayA-resilient";
</script>

<template>
  <div class="about">
    <v-sheet color="surface-container-low" rounded="xl" class="pa-6 pa-sm-8">
      <h1 class="md3-headline-medium mb-2">x-Ray VPN</h1>
      <p class="md3-body-large mb-1">{{ t("about.intro") }}</p>
      <p class="md3-body-medium mb-1">{{ t("about.protocols") }}</p>
      <p v-if="store.version" class="md3-body-medium text-on-surface-variant">
        {{ t("welcome.default", { version: store.version.version }) }}
      </p>
      <p class="md3-body-medium text-on-surface-variant mt-2">
        {{ t("about.founded") }}
        {{ t("about.local") }}
        {{ t("about.report", { discussions: t("about.discussions") }) }}
      </p>
      <p class="md3-body-medium mt-4">{{ t("about.experimental") }}</p>
      <p class="md3-body-medium mt-2">{{ t("about.fixes") }}</p>
      <p class="md3-body-medium text-on-surface-variant mt-2">
        {{ t("about.limitations") }}
      </p>
      <div class="d-flex flex-wrap ga-2 mt-6">
        <v-btn
          variant="tonal"
          :href="`${repo}/tree/experimental`"
          target="_blank"
          rel="noreferrer"
          :prepend-icon="mdiGithub"
          >GitHub</v-btn
        >
        <v-btn
          variant="tonal"
          href="https://github.com/v2rayA/v2rayA"
          target="_blank"
          rel="noreferrer"
          :prepend-icon="mdiGithub"
          >v2rayA</v-btn
        >
        <v-btn
          variant="tonal"
          href="#docs"
          :prepend-icon="mdiBookOpenPageVariantOutline"
          >{{ t("common.docs") }}</v-btn
        >
        <v-btn
          variant="tonal"
          :href="`${repo}/discussions`"
          target="_blank"
          rel="noreferrer"
          :prepend-icon="mdiForumOutline"
          >{{ t("about.discussions") }}</v-btn
        >
      </div>
    </v-sheet>
  </div>
</template>

<style scoped>
.about {
  overflow-wrap: anywhere;
}
</style>
