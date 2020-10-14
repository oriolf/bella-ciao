<script lang="ts">
  import Loading from "./Loading.svelte";
  import Alert from "./Alert.svelte";
  import Pagination from "./Pagination.svelte";
  import Button from "./Buttons/Button.svelte";
  import DownloadFileButton from "./Buttons/DownloadFileButton.svelte";
  import DeleteFileButton from "./Buttons/DeleteFileButton.svelte";
  import { get, submitFormJSON } from "../util";
  import type { User } from "../types/models.type";
  import { _ } from "svelte-i18n";

  export let url: string;
  export let error: string;
  export let unvalidated: boolean;
  let response: Promise<{
    users: User[],
    total: number
  }>;
  let page: number = 1;
  let itemsPerPage: number = 10;
  let query: string = "";
  let userMessageID: number;
  let userMessage: string = "";
  let timer;
  const JQ: any = jQuery;

  const debounce = (v) => {
    clearTimeout(timer);
    timer = setTimeout(() => {
      query = v;
    }, 400);
  };

  $: getUsers(page, query);

  async function getUsers(pg: number, qry: string) {
    response = get(
      `${url}?page=${pg}&items_per_page=${itemsPerPage}&query=${qry}`, null
    );
  }

  async function validateUser(id: number) {
    await fetch(`/api/users/validate?id=${id}`);
    getUsers(page, query);
  }

  function beginAddMessage(userID: number) {
    userMessageID = userID;
    JQ("#addMessageModal").on("shown.bs.modal", function () {
      JQ("#userMessageInput").trigger("focus");
    });
  }

  async function addMessage() {
    await submitFormJSON("/api/users/messages/add", {
      user_id: userMessageID,
      content: userMessage,
    }, null);
    JQ("#addMessageModal").modal("hide");
    userMessage = "";
    getUsers(page, query);
  }

  function debounceInput(event: Event) {
    const target = event.target as HTMLInputElement;
    debounce(target.value);
  }
</script>

<style>
  .user-container {
    border: 1px solid rgba(0, 0, 0, 0.125);
    border-radius: 3px;
    margin-bottom: 5px;
  }
</style>

<div class="card">
  <div class="card-body">
    <div class="input-group mb-3">
      <input
        type="text"
        class="form-control"
        placeholder={$_("comp.users_pagination.filter_user_id")}
        aria-label={$_("comp.users_pagination.filter_user_id")}
        aria-describedby="filter-users"
        on:keyup={debounceInput} />
    </div>
    {#await response}
      <Loading />
    {:then resp}
      {#each resp.users || [] as user}
        <div class="container user-container">
          <div class="row" style="margin-top: 5px;">
            <div class="col-6">
              <h6>
                {user.unique_id}
                {#if user.role === 'admin'}<em>(admin)</em>{/if}
              </h6>
              <p>{user.name} <em>({user.email})</em></p>
            </div>
            {#if unvalidated}
              <div class="col-3">
                <button
                  type="button"
                  class="align-middle btn btn-sm btn-outline-primary"
                  style="width: 100%;"
                  data-toggle="modal"
                  data-target="#addMessageModal"
                  on:click={() => beginAddMessage(user.id)}>
                  {$_("comp.users_pagination.add_message")}
                </button>
              </div>
              <div class="col-3">
                <Button
                  content="Validate"
                  type="success"
                  callback={() => validateUser(user.id)} />
              </div>
            {/if}
          </div>

          <div class="row">
            <div class="col-12">
              <h6>{$_("comp.users_pagination.files")}</h6>
            </div>
          </div>

          {#each user.files || [] as file}
            <div class="row" style="margin-top: 2px;">
              <div class="col-6" style="padding-left: 32px;">
                {file.description}
              </div>
              <div class="col-3">
                <DownloadFileButton id={file.id} filename={file.name} />
              </div>
              <div class="col-3">
                <DeleteFileButton
                  id={file.id}
                  on:executed={() => getUsers(page, query)} />
              </div>
            </div>
          {/each}

          <div class="row">
            <div class="col-12">
              <h6>{$_("comp.users_pagination.messages")}</h6>
            </div>
          </div>

          {#each user.messages || [] as message}
            <div class="row" style="margin: 2px 0 0 5px;">
              <div class="col-12">
                {#if message.solved}
                  <span class="badge badge-primary">{$_("comp.users_pagination.solved")}</span>
                {:else}
                  <span class="badge badge-warning">{$_("comp.users_pagination.not_solved")}</span>
                {/if}
                {message.content}
              </div>
            </div>
          {/each}
        </div>
      {/each}
      <Pagination bind:page {itemsPerPage} totalItems={resp.total} />
    {:catch _}
      <Alert type="danger" content={error} />
    {/await}
  </div>
</div>

<!-- Modal -->
<div
  class="modal fade"
  id="addMessageModal"
  tabindex="-1"
  role="dialog"
  aria-labelledby="addMessageModalLabel"
  aria-hidden="true">
  <div class="modal-dialog" role="document">
    <div class="modal-content">
      <div class="modal-header">
        <h5 class="modal-title" id="addMessageModalLabel">{$_("comp.users_pagination.add_message")}</h5>
      </div>
      <div class="modal-body">
        <div class="form-group">
          <textarea
            id="userMessageInput"
            class="form-control"
            rows="3"
            bind:value={userMessage} />
        </div>
      </div>
      <div class="modal-footer">
        <button
          type="button"
          class="btn btn-secondary"
          data-dismiss="modal">{$_("comp.users_pagination.close")}</button>
        <button type="button" class="btn btn-primary" on:click={addMessage}>
          {$_("comp.users_pagination.add_message")}
        </button>
      </div>
    </div>
  </div>
</div>
