<script lang="ts">
  import Alert from "./Alert.svelte";
  import Button from "./Buttons/Button.svelte";
  import CardTable from "./CardTable.svelte";
  import Form from "./Form.svelte";
  import UserFiles from "./UserFiles.svelte";
  import { get, sortByField } from "../util";
  import type { UserMessage, FormParams } from "../types/models.type";
  import { _ } from "svelte-i18n";

  let messages: Promise<UserMessage[]>;
  let reloadFiles: number = 0;
  let uploadFileForm: FormParams = {
    name: "comp.role_none.upload_file",
    values: "form",
    url: "/api/users/files/upload",
    generalError: "comp.role_none.upload_file_err",
    fields: [
      {
        name: "description",
        hint: "comp.role_none.description_hint",
        required: true,
      },
      {
        name: "file",
        hint: "comp.role_none.upload_file",
        required: true,
        type: "file",
      },
    ],
  };

  getMessages();

  async function getMessages() {
    messages = get("/api/users/messages/own", sortByField("solved"));
  }

  // TODO handle errors
  async function solveMessage(id: number) {
    await fetch(`/api/users/messages/solve?id=${id}`);
    getMessages();
  }
</script>

<Alert content={$_("comp.role_none.none_notice")} />

<h2>{$_("comp.role_none.messages_title")}</h2>
<p>{$_("comp.role_none.messages_explanation")}</p>

<CardTable
  headers={[$_("comp.role_none.messages_header"), '', '']}
  rows={messages}
  let:row
  error={$_("comp.role_none.messages_err")}>
  <td>{row.content}</td>
  <td>
    {#if row.solved}
      <em>{$_("comp.role_none.solved")}</em>
    {:else}
      <Button content={$_("comp.role_none.solve")} callback={() => solveMessage(row.id)} />
    {/if}
  </td>
</CardTable>

<h2>{$_("comp.role_none.files_title")}</h2>
<p>{$_("comp.role_none.files_explanation")}</p>

<UserFiles {reloadFiles} />

<div class="card">
  <div class="card-header">{$_("comp.role_none.upload_file")}</div>
  <div class="card-body">
    <Form params={uploadFileForm} on:executed={() => (reloadFiles += 1)} />
  </div>
</div>
