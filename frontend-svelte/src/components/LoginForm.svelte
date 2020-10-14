<script lang="ts">
  import { createEventDispatcher } from "svelte";
  import Form from "./Form.svelte";
  import type { FormParams } from "../types/models.type";
  import { _ } from "svelte-i18n";

  const dispatch = createEventDispatcher();
  let loggingIn: boolean = true;

  const loginForm: FormParams = {
    name: "comp.login_form.login",
    url: "/api/auth/login",
    generalError: "comp.login_form.login_err",
    fields: [
      {
        name: "unique_id",
        hint: "comp.login_form.unique_id_hint",
        required: true,
        errString: "comp.login_form.unique_id_err"
      },
      {
        name: "password",
        hint: "comp.login_form.password_hint",
        required: true,
        type: "password",
        errString: "comp.login_form.password_err"
      },
    ],
  };

  const registerForm: FormParams = {
    name: "comp.login_form.register",
    url: "/api/auth/register",
    generalError: "comp.login_form.register_err",
    fields: [
      {
        name: "unique_id",
        hint: "comp.login_form.unique_id_hint",
        required: true,
        errString: "comp.login_form.unique_id_err"
      },
      {
        name: "name",
        hint: "comp.login_form.name_hint",
        required: true,
      },
      {
        name: "email",
        hint: "comp.login_form.email_hint",
        required: true,
        type: "email",
      },
      // TODO validate passwords match
      {
        name: "password",
        hint: "comp.login_form.password_hint",
        required: true,
        type: "password",
        errString: "comp.login_form.register_password_err"
      },
      {
        name: "repeat_password",
        hint: "comp.login_form.repeat_password_hint",
        required: true,
        type: "password",
        errString: "comp.login_form.register_password_err"
      },
    ],
  };
</script>

{#if loggingIn}
  <div class="card">
    <div class="card-header">Log in</div>
    <div class="card-body">
      <Form params={loginForm} on:executed={(_) => dispatch('loggedin', true)}>
        <a
          href="."
          class="btn"
          on:click={(_) => {
            loggingIn = false;
          }}>
          {$_("comp.login_form.register")}
        </a>
      </Form>
    </div>
  </div>
{:else}
  <div class="card">
    <div class="card-header">{$_("comp.login_form.register")}</div>
    <div class="card-body">
      <Form
        params={registerForm}
        on:executed={(_) => {
          loggingIn = true;
        }}>
        <a
          href="."
          class="btn"
          on:click={(_) => {
            loggingIn = true;
          }}>
          {$_("comp.login_form.login")}
        </a>
      </Form>
    </div>
  </div>
{/if}
