<script>
  import { user } from "../store";
  import { whoami } from "../util";
  import CandidatesList from "../components/CandidatesList.svelte";
  import CandidateForm from "../components/CandidateForm.svelte";

  let isAdmin = false;
  let reloadCandidates = 0;
  getUser();

  async function getUser() {
    await whoami(user);
    isAdmin = $user && $user.role === "admin";
  }
</script>

<CandidatesList admin={isAdmin} bind:reloadCandidates />

{#if isAdmin}
  <CandidateForm on:executed={() => {reloadCandidates += 1}} />
{/if}
