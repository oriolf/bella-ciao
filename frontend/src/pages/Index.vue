<template>
  <q-page class="flex">
    <div class="row" style="width: 100%;">
      <div class="col" style="margin: 10px;">
        <div class="text-h6">{{ $t("ACTIVE_VOTES") }}</div>

        <q-list bordered class="rounded-borders">
          <q-expansion-item
            v-for="election in elections"
            :key="election.id"
            expand-separator
            :label="electionLabel(election)"
            :caption="electionCaption(election)"
          >
            <q-card>
              <q-card-section>
                <ul>
                  <li
                    v-for="candidate in election.candidates"
                    :key="candidate.id"
                  >
                    {{ candidate.name }}
                  </li>
                </ul>
              </q-card-section>
            </q-card>
          </q-expansion-item>
        </q-list>
      </div>

      <div v-if="!$token.value" class="col" style="margin: 10px;">
        <q-card>
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

      <div v-else class="col" style="margin: 10px;">
        <p v-if="$token.value.role === 'admin'">You are an administrator</p>
        <p v-else-if="$token.value.role === 'validated'">
          You have been validated
        </p>
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
      this.$axios.defaults.headers.common["Authorization"] =
        "bearer " + tokenStr;
      this.reloadData();
    },
    logout: function() {
      this.$token.value = null;
      this.$token.str = "";
      this.$axios.defaults.headers.common["Authorization"] = "";
      this.$router.push("/");
      this.$forceUpdate();
    }
  }
};

const apiURL = suffix => {
  let location = window.location;
  return location.protocol + "//" + location.hostname + ":9876" + suffix;
};

export default {
  name: "PageIndex",
  mixins: [tokenMixin],
  data: function() {
    return {
      uniqueID: "",
      password: "",
      elections: null
    };
  },
  created: function() {
    this.reloadData();
  },
  methods: {
    reloadData: function() {
      this.reloadElections();
    },
    reloadElections: function() {
      this.$axios
        .get(apiURL("/elections/get"))
        .then(res => {
          this.elections = res.data;
        })
        .catch(() => {
          this.$q.notify("Error getting elections");
        });
    },
    onLogin: function() {
      this.$axios
        .post(
          apiURL("/auth/login"),
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
    },
    electionCaption: function(election) {
      let from = this.$t("FROM_DATE");
      let to = this.$t("TO_DATE");
      let start = new Date(election.start).toLocaleString();
      let end = new Date(election.end).toLocaleString();
      return `${from} ${start} ${to} ${end}`;
    },
    electionLabel: function(election) {
      if (!election.public) {
        let isPrivate = this.$t("PRIVATE");
        return election.name + ` (${isPrivate})`;
      }
      return election.name;
    }
  }
};
</script>
