describe('Admin user', () => {
    before(() => {
        cy.exec("npm run db:validate");
    });

    beforeEach(() => {
        cy.visit('/', { onBeforeLoad: () => { localStorage.setItem('bella-ciao.lang', 'en') } });
        cy.get('#unique_id').type("11111111H");
        cy.get('#password').type("12345678");
        cy.get('button').contains('Log in').click();
    });

    it('shows the list of existing users and allows validation', () => {
        cy.get("#validated").contains('h6', '11111111H');
        cy.get("#validated").contains('h6', '22222222J');
        cy.get("#validated").contains('h6', '33333333P');
        cy.get("#validated").contains('h6', 'X1111111G');

        cy.get("#unvalidated").contains('h6', '88888888Y');

        cy.get("#unvalidated").get("button").contains("Validate").click();

        cy.get("#unvalidated").contains('h6', '88888888Y').should("not.exist");
    });
});