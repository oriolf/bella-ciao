<script>
  import UserFiles from "./UserFiles.svelte";
  import Alert from "./Alert.svelte";
  import { get, formatDate } from "../util.js";

  let promise;
  getElection();
  function getElection() {
    promise = get("/api/elections/get");
  }
</script>

{#await promise}
  <Alert content="You have been validated" />
{:then elections}
  <Alert
    content="You have been validated. The election will take place between {formatDate(elections[0].start)}
    and {formatDate(elections[0].end)}" />
{:catch _}
  <Alert content="You have been validated" />
{/await}

<h2>Uploaded files</h2>

<UserFiles />
