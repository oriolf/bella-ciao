<template>
  <q-layout view="lHh Lpr lFf">
    <q-header elevated>
      <q-toolbar class="bg-purple text-white">
        <q-btn
          flat
          round
          aria-label="Menu"
          @click="leftDrawerOpen = !leftDrawerOpen"
        >
          <q-icon name="menu" />
        </q-btn>

        <q-toolbar-title>Bella Ciao</q-toolbar-title>

        <span v-if="$token.value">Welcome {{ $token.value.name }}</span>
        <q-btn v-if="$token.value" flat round @click="onLogout">
          <q-icon name="logout" />
        </q-btn>
      </q-toolbar>
    </q-header>

    <q-drawer v-model="leftDrawerOpen" bordered content-class="bg-grey-2">
      <q-list>
        <q-item-label header>{{ $t("MENU") }}</q-item-label>
        <q-item to="/">
          <q-item-section avatar>
            <q-icon name="home" />
          </q-item-section>
          <q-item-section>
            <q-item-label>{{ $t("HOME") }}</q-item-label>
          </q-item-section>
        </q-item>

        <q-item to="/faqs">
          <q-item-section avatar>
            <q-icon name="question_answer" />
          </q-item-section>
          <q-item-section>
            <q-item-label>{{ $t("faqs.title") }}</q-item-label>
          </q-item-section>
        </q-item>
      </q-list>
    </q-drawer>

    <q-page-container>
      <router-view />
    </q-page-container>
  </q-layout>
</template>

<script>
import tokenMixin from "../pages/Index";

export default {
  name: "MyLayout",
  mixins: [tokenMixin],
  data() {
    return {
      leftDrawerOpen: this.$q.platform.is.desktop
    };
  },
  methods: {
    onLogout: function() {
      this.logout();
    }
  }
};
</script>

<style></style>
