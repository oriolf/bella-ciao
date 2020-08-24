<script>
  import { createEventDispatcher } from "svelte";
  import Form from "./Form.svelte";

  const dispatch = createEventDispatcher();
  let loggingIn = true;

  const loginForm = {
    name: "Log in",
    url: "/api/auth/login",
    generalError: "The user or the password are invalid",
    fields: [
      {
        name: "unique_id",
        hint: "Enter user (identification number: DNI, NIE, etc.)",
        required: true,
        errString: "This field is required and must match a valid ID number format"
      },
      {
        name: "password",
        hint: "Password",
        required: true,
        type: "password",
        errString: "This field is required and must be eight or more characters long"
      },
    ],
  };

  const registerForm = {
    name: "Register",
    url: "/api/auth/register",
    generalError:
      "The user is not a valid identification number (DNI, NIE...), the name or email are empty, or the passwords don't match",
    fields: [
      {
        name: "unique_id",
        hint: "Enter user (identification number: DNI, NIE, etc.)",
        required: true,
        errString: "This field is required and must match a valid ID number format"
      },
      {
        name: "name",
        hint: "Enter your name",
        required: true,
      },
      {
        name: "email",
        hint: "Enter your email",
        required: true,
        type: "email",
      },
      // TODO validate passwords match
      {
        name: "password",
        hint: "Password",
        required: true,
        type: "password",
        errString: "This field is required and must match the repeat password field"
      },
      {
        name: "repeat_password",
        hint: "Repeat password",
        required: true,
        type: "password",
        errString: "This field is required and must match the password field"
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
          Register
        </a>
      </Form>
    </div>
  </div>
{:else}
  <div class="card">
    <div class="card-header">Register</div>
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
          Log in
        </a>
      </Form>
    </div>
  </div>
{/if}
