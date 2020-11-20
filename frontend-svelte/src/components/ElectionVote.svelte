<script lang="ts">
  import { _ } from "svelte-i18n";
  import { user } from "../store";
  import { formatDate, post } from "../util";
  import Alert from "./Alert.svelte";
  import Form from "./Form.svelte";
  import CandidatesVoteList from "./CandidatesVoteList.svelte";
  import ElectionResults from "./ElectionResults.svelte";
  import Button from "./Buttons/Button.svelte";
  import type { Election, Candidate } from "../types/models.type";

  export let election: Election;
  let unselectedCandidates: Candidate[] = election.candidates;
  let selectedCandidates: Candidate[] = [];
  let candidatesVotedFor: Candidate[];
  let disableVote = true;
  let voteHash: string;

  $: disableVote =
    selectedCandidates.length < election.min_candidates ||
    selectedCandidates.length > election.max_candidates;

  // TODO show error if voted hash not found

  function select(event) {
    [unselectedCandidates, selectedCandidates] = move(
      event.detail,
      unselectedCandidates,
      selectedCandidates
    );
  }

  function deselect(event) {
    [selectedCandidates, unselectedCandidates] = move(
      event.detail,
      selectedCandidates,
      unselectedCandidates
    );
  }

  function move(
    id: number,
    l1: Candidate[],
    l2: Candidate[]
  ): [Candidate[], Candidate[]] {
    let candidate = l1.filter((x) => x.id === id)[0];
    l2.push(candidate);
    l2 = l2;
    l1 = l1.filter((x) => x.id !== id);
    return [l1, l2];
  }

  function up(event) {
    let id = event.detail;
    let index = selectedCandidates.findIndex((x) => x.id === id);
    if (index === 0) {
      return;
    }
    let candidate = selectedCandidates.splice(index, 1)[0];
    selectedCandidates.splice(index - 1, 0, candidate);
    selectedCandidates = selectedCandidates;
  }

  function down(event) {
    let id = event.detail;
    let index = selectedCandidates.findIndex((x) => x.id === id);
    if (index === selectedCandidates.length - 1) {
      return;
    }
    let candidate = selectedCandidates.splice(index, 1)[0];
    selectedCandidates.splice(index + 1, 0, candidate);
    selectedCandidates = selectedCandidates;
  }

  async function vote() {
    let res = await post("/api/elections/vote", {
      candidates: selectedCandidates.map((x) => x.id),
    });
    voteHash = await res.json();
    $user.has_voted = true;
  }

  const checkVoteForm = {
    name: "comp.election_vote.form_name",
    url: "/api/elections/vote/check",
    fields: [
      {
        name: "token",
        hint: "comp.election_vote.token",
        required: true,
      },
    ],
  };
</script>

<h2>{$_("comp.election_vote.election")}</h2>

{#if new Date() < new Date(election.start)}
  <p>
    {$_("comp.election_vote.election_before", {
      values: { start: formatDate(election.start), end: formatDate(election.end) }
    })}
  </p>
{:else if new Date() > new Date(election.start) && new Date() < new Date(election.end)}
  <p>
    {$_('comp.election_vote.election_started', {
      values: { end: formatDate(election.end) },
    })}
  </p>
  {#if $user.has_voted}
    {#if voteHash}
      <Alert
        content={$_("comp.election_vote.vote_confirmation", { values: { hash: voteHash } })} />
    {:else}
      <Alert content={$_("comp.election_vote.already_voted")} type="warning" />
      <div class="card" style="padding: 5px;">
        <div class="row">
          <div class="col-12">
            {#if !candidatesVotedFor}
              <p>{$_("comp.election_vote.check_vote")}</p>
              <Form
                params={checkVoteForm}
                on:result={(res) => {
                  candidatesVotedFor = res.detail;
                }} />
            {:else}
              <p>{$_("comp.election_vote.candidates_list")}</p>
              <ol>
                {#each candidatesVotedFor as candidate}
                  <li>{candidate.name}</li>
                {/each}
              </ol>
            {/if}
          </div>
        </div>
      </div>
    {/if}
  {:else}
    <div class="card" id="candidates" style="padding: 5px;">
      <div class="row">
        <div class="col-12">
          <p>{$_("comp.election_vote.candidates_range", {values: { min: election.min_candidates, max: election.max_candidates }})}</p>
        </div>
      </div>
      <div class="row">
        <div class="col-6">
          <h2>{$_("comp.election_vote.candidates")}</h2>

          <CandidatesVoteList
            selected={false}
            list={unselectedCandidates}
            on:select={select} />
        </div>
        <div class="col-6">
          <h2>{$_("comp.election_vote.selected_candidates")}</h2>
          <CandidatesVoteList
            selected={true}
            list={selectedCandidates}
            on:deselect={deselect}
            on:up={up}
            on:down={down} />
        </div>
      </div>
      <div>
        <Button content={$_("comp.election_vote.vote")} callback={vote} disabled={disableVote} />
      </div>
    </div>
  {/if}
{:else}
  <p>{$_("comp.election_vote.election_results")}</p>
  <ElectionResults {election} />
{/if}
