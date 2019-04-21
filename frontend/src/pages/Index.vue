<template>
  <q-page class="flex">
    <div class="row" style="width: 100%;">
      <div class="col"></div>

      <div v-if="!$token.value" class="col">
        <q-card style="margin: 10px;">
          <q-card-section>
            <div class="text-h6">Login</div>
          </q-card-section>

          <q-card-section>
            <q-form class="q-gutter-md" @submit="onLogin" @reset="onResetLogin">
              <q-input
                v-model="uniqueID"
                filled
                label="Your unique ID"
                lazy-rules
                :rules="[
                  val => (val && val.length > 0) || 'Please type something'
                ]"
              />

              <q-input
                v-model="password"
                filled
                type="password"
                label="Password"
                lazy-rules
                :rules="[
                  val => (val && val.length > 5) || 'Please type something'
                ]"
              />

              <div>
                <q-btn label="Log in" type="submit" color="primary" />
              </div>
            </q-form>
          </q-card-section>
        </q-card>
      </div>

      <div v-else class="col">
        <p v-if="$token.value.role === 'admin'">You are an administrator</p>
        <p v-if="$token.value.role === 'validated'">You have been validated</p>
        <p v-else>You have not been validated yet</p>
      </div>
    </div>
  </q-page>
</template>

<style></style>

<script>
// TODO proper and translated messages for login form
// TODO if have a valid token, show user info on top, and substitute login form by internal info
export const tokenMixin = {
  methods: {
    updateToken: function(tokenStr) {
      var base64Url = tokenStr.split(".")[1];
      var base64 = base64Url.replace(/-/g, "+").replace(/_/g, "/");
      this.$token.value = JSON.parse(window.atob(base64));
      this.$token.str = tokenStr;
      this.$router.push("/"); // TODO maybe not always reload...
      this.$forceUpdate(); // TODO does not work
    },
    logout: function() {
      this.$token.value = null;
      this.$token.str = "";
      this.$router.push("/");
      this.$forceUpdate();
    }
  }
};

export default {
  name: "PageIndex",
  mixins: [tokenMixin],
  data: function() {
    return {
      uniqueID: "",
      password: ""
    };
  },
  methods: {
    onLogin: function() {
      this.$axios
        .post(
          // TODO proper url handling
          "http://localhost:9876/auth/login",
          JSON.stringify({
            unique_id: this.uniqueID,
            password: this.password
          })
        )
        .then(response => this.updateToken(response.data))
        .catch(() =>
          this.$q.notify("Error logging in, inexistent user or wrong password")
        );
    },
    onResetLogin: function() {
      this.uniqueID = "";
      this.password = "";
    }
  }
};
</script>
