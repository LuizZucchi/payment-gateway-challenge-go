# Design Decisions

## Architectural Decision: Package by Feature

- Decision: Refactored the project structure from technical layers (e.g., handlers/, repository/, models/) to domain-centric packages (e.g., internal/payments/).

- Rationale:
    - High Cohesion: All logic related to payments (entities, storage, validation, transport) is located in a single place, making the code easier to navigate and understand.
    - Encapsulation: Allows better control over what is exported publicly vs. what is internal to the payment domain.
    - Scalability: Facilitates future extraction into microservices if the domain grows, as the dependencies are already isolated by feature.
    - Customization: not every package has to follow a tight layered approach, this enable us to allow for complexity to grow with our needs
