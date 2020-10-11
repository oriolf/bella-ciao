<script>
  // TODO convert all to TypeScript
  import UsersPagination from "./UsersPagination.svelte";
  import Loading from "./Loading.svelte";
  import ElectionVote from "./ElectionVote.svelte";
  import { get } from "../util";

  let promise;
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
  error="Could not get users pending validation"
  url="/api/users/unvalidated/get" />

<h2>Validated users</h2>

<UsersPagination
  unvalidated={false}
  error="Could not get validated users"
  url="/api/users/validated/get" />
