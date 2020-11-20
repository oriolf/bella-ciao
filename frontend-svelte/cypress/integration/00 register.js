// TODO register user, login, send some file
describe('User registration', () => {
    before(() => {
        cy.exec("npm run db:register");
    });

    it('can submit the registration form', () => {
        cy.visit('/', { onBeforeLoad: () => { localStorage.setItem('bella-ciao.lang', 'en') } });

        cy.contains("Register").click();
        cy.get("[placeholder^='Enter user']").type("88888888Y");
        cy.get("[placeholder^='Enter your name']").type("User 8 name");
        cy.get("[placeholder^='Enter your email']").type("name8@example.com");
        cy.get("[placeholder^='Password']").type("12345678");
        cy.get("[placeholder^='Repeat password']").type("12345678");
        cy.get("button").contains("Register").click();
    });

    it('can log in after registering and upload a file', () => {
        cy.visit('/', { onBeforeLoad: () => { localStorage.setItem('bella-ciao.lang', 'en') } });

        cy.get("[placeholder^='Enter user']").type("88888888Y");
        cy.get("[placeholder^='Password']").type("12345678");
        cy.get("button").contains("Log in").click();

        cy.contains(".alert", "You still have not been validated");
        cy.contains(".card-header", "Upload file");

        const uploadedFile = "Uploaded file name";
        cy.get("[placeholder^='Description of the file']").type(uploadedFile);
        cy.fixture('file.jpg').then(fileContent => {
            cy.get('input[type="file"]').attachFile({
                fileContent: fileContent.toString(),
                fileName: 'file.jpg',
                mimeType: 'image/jpg'
            });
        });
        cy.get("button").contains("Upload").click();

        cy.get(".card table tbody").contains(uploadedFile);
    });
});