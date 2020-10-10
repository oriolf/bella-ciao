<script lang="ts">
  import UserFiles from "./UserFiles.svelte";
  import Alert from "./Alert.svelte";
  import Loading from "./Loading.svelte";
  import ElectionVote from "./ElectionVote.svelte";
  import { get } from "../util";
  import type { Election } from "../types/models.type";

  let promise: Promise<Election[]>;
  getElection();
  function getElection() {
    promise = get("/api/elections/get", null);
  }
</script>

<Alert content="You have been validated" />
{#await promise}
  <Loading />
{:then elections}
  <ElectionVote election={elections[0]} />
{/await}

<h2>Uploaded files</h2>

<UserFiles />
