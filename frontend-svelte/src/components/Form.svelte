<script>
  import { submitForm, validationFuncs, extractFormValues } from "../util";
  import { createEventDispatcher, onMount } from "svelte";

  export let params;
  const dispatch = createEventDispatcher();
  let errors = {};
  let fields = {};
  let generalError = "";
  let valFuncs;

  onMount(() => {
    let l = [];
    for (let i = 0; i < params.fields.length; i++) {
      let field = params.fields[i];
      fields[field.name] = field;
      if (field.validate) {
        l.push(field.validate(field.name));
      }
      if (field.required && !field.errString) {
        params.fields[i].errString = "This field is required";
      }
    }

    valFuncs = validationFuncs(l);
  });

  async function submit(event) {
    event.preventDefault();
    const values = extractFormValues(event);
    errors = valFuncs(values);

    // TODO update error messages on each change in inputs, after the first submit attempt
    let form = event.target;
    for (let el of form.elements) {
      if (errors[el.name]) {
        el.setCustomValidity(errors[el.name]);
      } else if (
        fields[el.name] &&
        fields[el.name].required &&
        el.value === ""
      ) {
        el.setCustomValidity("This field is required");
        errors[el.name] = "This field is required";
      } else {
        el.setCustomValidity("");
      }
    }

    form.classList.add("was-validated");
    if (!form.checkValidity()) {
      return;
    }

    const res = await submitForm(params.url, values, params.jsonFunc);
    if (res.ok) {
      dispatch("executed", true);
    } else {
      if (Object.keys(errors).length === 0) {
        generalError = params.generalError;
      }
    }
  }
</script>

{#if !params}
  <p>Undefined form</p>
{:else}
  <form on:submit={submit} class="needs-validation" novalidate>
    {#each params.fields as field}
      <div class="form-group">
        <input
          class="form-control"
          class:is-invalid={errors[field.name]}
          id={field.name}
          name={field.name}
          aria-describedby={field.name + '_help'}
          placeholder={field.hint}
          type={field.type || 'text'}
          required={field.required} />
        <div class="invalid-feedback">{field.errString}</div>
      </div>
    {/each}
    <button type="submit" class="btn btn-primary">{params.name}</button>
    <slot />
    {#if generalError}
      <div class="alert alert-danger" role="alert" style="margin: 15px 0 0 0;">
        {generalError}
      </div>
    {/if}
  </form>
{/if}
