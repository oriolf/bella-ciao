<script lang="ts">
  import UsersPagination from "./UsersPagination.svelte";
  import Loading from "./Loading.svelte";
  import ElectionVote from "./ElectionVote.svelte";
  import { get } from "../util";
  import type { Election } from "../types/models.type";
  import { _ } from "svelte-i18n";

  let promise: Promise<Election[]>;
  getElection();
  function getElection() {
    promise = get("/api/elections/get", null);
  }
</script>

{#await promise}
  <Loading />
{:then elections}
  <ElectionVote election={elections[0]} />
{/await}

<h2>Users pending validation</h2>

<UsersPagination
  unvalidated={true}
  error={$_("comp.role_admin.pending_err")}
  url="/api/users/unvalidated/get" />

<h2>Validated users</h2>

<UsersPagination
  unvalidated={false}
  error={$_("comp.role_admin.validated_err")}
  url="/api/users/validated/get" />
