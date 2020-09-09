<script>
  import { get } from "../util.js";
  import Alert from "./Alert.svelte";
  import Loading from "./Loading.svelte";
  import Button from "./Buttons/Button.svelte";

  let promise;
  export let reloadCandidates;
  export let admin;

  $: getCandidates(reloadCandidates);

  getCandidates();
  function getCandidates() {
    promise = get("/api/candidates/get");
  }

  async function deleteCandidate(id) {
    await fetch(`/api/candidates/delete?id=${id}`);
    getCandidates();
  }
</script>

{#await promise}
  <Loading />
{:then candidates}
  {#each candidates as candidate}
    <div class="card" style="margin-bottom: 20px;">
      <div class="card-header">
        {candidate.name}{#if admin}
          <span class="float-right"><Button
              content="Delete candidate"
              type="danger"
              callback={() => deleteCandidate(candidate.id)} /></span>
        {/if}
      </div>
      <div class="card-body">
        <div class="row">
          <div class="col-md-3">
            <img
              alt={`Candidate ${candidate.name} image`}
              style="max-width: 100%;"
              src={`/api/candidates/image?id=${candidate.id}`} />
          </div>
          <div class="col-md-9">{candidate.presentation}</div>
        </div>
      </div>
    </div>
  {/each}
{:catch _}
  <Alert type="danger" content="Could not get candidates" />
{/await}
