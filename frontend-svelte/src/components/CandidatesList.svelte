<script lang="ts">
  import { get } from "../util";
  import Alert from "./Alert.svelte";
  import Loading from "./Loading.svelte";
  import Button from "./Buttons/Button.svelte";
  import type { Candidate } from "../types/models.type";
  import { _ } from "svelte-i18n";

  let promise: Promise<Candidate[]>;
  export let reloadCandidates: number;
  export let admin: boolean;

  $: getCandidates(reloadCandidates);

  function getCandidates(_: number) {
    promise = get("/api/candidates/get", null) as Promise<Candidate[]>;
  }

  async function deleteCandidate(id: number) {
    await fetch(`/api/candidates/delete?id=${id}`);
    getCandidates(reloadCandidates);
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
              content={$_("comp.candidates_list.delete")}
              type="danger"
              callback={() => deleteCandidate(candidate.id)} /></span>
        {/if}
      </div>
      <div class="card-body">
        <div class="row">
          <div class="col-md-3">
            <img
              alt={$_("comp.candidates_list.image", { values: { name: candidate.name } })}
              style="max-width: 100%;"
              src={`/api/candidates/image?id=${candidate.id}`} />
          </div>
          <div class="col-md-9">{candidate.presentation}</div>
        </div>
      </div>
    </div>
  {/each}
{:catch e}
  <Alert type="danger" content={$_("comp.candidates_list.error")} />
{/await}
