<script>
  import {
    submitForm,
    submitFormJSON,
    validationFuncs,
    extractFormValuesJSON,
  } from "../util";
  import { createEventDispatcher, onMount } from "svelte";

  export let params;
  const dispatch = createEventDispatcher();
  let errors = {};
  let fields = {};
  let generalError = "";
  let valFuncs;
  let wasValidated = false;
  let multiselectFields;

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
    multiselectFields = params.fields
      .filter((x) => x.type === "multiselect")
      .map((x) => x.name);
  });

  function updateValidation(event) {
    let form = event.target.form;
    const values = extractFormValuesJSON(form, multiselectFields);
    errors = valFuncs(values);
    updateValidationAux(form, values, errors);
  }

  function updateValidationAux(form, values, errors) {
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

    return form.checkValidity();
  }

  async function submit(event) {
    event.preventDefault();

    let form = event.target;
    const values = extractFormValuesJSON(form, multiselectFields);
    errors = valFuncs(values);
    form.classList.add("was-validated");
    wasValidated = true;
    if (!updateValidationAux(form, values, errors)) {
      return;
    }

    let res;
    if (params.values === "form") {
      res = await submitForm(params.url, event.target);
    } else {
      res = await submitFormJSON(params.url, values, params.jsonFunc);
    }
    if (res.ok) {
      dispatch("executed", true);
      form.reset();
      form.classList.remove("was-validated");
      wasValidated = false;
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
      {#if field.type === 'file'}
        <div class="form-group row" style="margin: 10px 0 10px 0;">
            <input
              on:input={updateValidation}
              type="file"
              class="form-control-file"
              id={field.name}
              name={field.name}
              required={field.required} />
            <div class="invalid-feedback">{field.errString}</div>
        </div>
      {:else if field.type === 'textarea'}
        <textarea
          on:input={updateValidation}
          class="form-control"
          class:is-invalid={wasValidated && errors[field.name]}
          id={field.name}
          name={field.name}
          placeholder={field.hint}
          required={field.required}
          rows="3" />
      {:else if field.type === 'datetime'}
        <div class="form-group row">
          {#if field.title}
            <label
              for={field.name + '_date'}
              class="col-md-3 col-form-label">{field.title} (date)</label>
          {/if}
          <div class={field.title ? 'col-md-9' : 'col-md-12'}>
            <input
              on:input={updateValidation}
              class="form-control"
              class:is-invalid={wasValidated && errors[field.name + '_date']}
              id={field.name + '_date'}
              name={field.name + '_date'}
              aria-describedby={field.name + '_help'}
              placeholder={field.hint}
              type="date"
              required={field.required} />
            <div class="invalid-feedback">{field.errString}</div>
          </div>
        </div>
        <div class="form-group row">
          {#if field.title}
            <label
              for={field.name + '_time'}
              class="col-md-3 col-form-label">{field.title} (time)</label>
          {/if}
          <div class={field.title ? 'col-md-9' : 'col-md-12'}>
            <input
              on:input={updateValidation}
              class="form-control"
              class:is-invalid={wasValidated && errors[field.name + '_time']}
              id={field.name + '_time'}
              name={field.name + '_time'}
              aria-describedby={field.name + '_help'}
              placeholder={field.hint}
              type="time"
              required={field.required} />
            <div class="invalid-feedback">{field.errString}</div>
          </div>
        </div>
      {:else if field.type === 'select'}
        <div class="form-group row">
          {#if field.title}
            <label
              for={field.name}
              class="col-md-3 col-form-label">{field.title}</label>
          {/if}
          <div class={field.title ? 'col-md-9' : 'col-md-12'}>
            <select
              on:input={updateValidation}
              class="form-control"
              class:is-invalid={wasValidated && errors[field.name]}
              id={field.name}
              name={field.name}
              aria-describedby={field.name + '_help'}
              required={field.required}>
              <option disabled selected value>{field.hint}</option>
              {#each field.options as option}
                <option value={option.id}>{option.name}</option>
              {/each}
            </select>
            <div class="invalid-feedback">{field.errString}</div>
          </div>
        </div>
      {:else if field.type === 'multiselect'}
        <div class="form-group row">
          {#if field.title}
            <label
              for={field.name}
              class="col-md-3 col-form-label">{field.title}</label>
          {/if}
          <div class={field.title ? 'col-md-9' : 'col-md-12'}>
            <div class="form-check form-check-inline">
              {#each field.options as option}
                <input
                  on:input={updateValidation}
                  class="form-check-input"
                  class:is-invalid={wasValidated && errors[field.name]}
                  name={field.name}
                  type="checkbox"
                  value={option.id} />
                <label
                  class="form-check-label"
                  style="margin-right: 15px;"
                  for={field.name}>{option.name}</label>
              {/each}
            </div>
          </div>
        </div>
      {:else}
        <div class="form-group row">
          {#if field.title}
            <label
              for={field.name}
              class="col-md-3 col-form-label">{field.title}</label>
          {/if}
          <div class={field.title ? 'col-md-9' : 'col-md-12'}>
            <input
              on:input={updateValidation}
              class="form-control"
              class:is-invalid={wasValidated && errors[field.name]}
              id={field.name}
              name={field.name}
              aria-describedby={field.name + '_help'}
              placeholder={field.hint}
              type={field.type || 'text'}
              required={field.required} />
            <div class="invalid-feedback">{field.errString}</div>
          </div>
        </div>
      {/if}
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
