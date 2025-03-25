# Software Requirements Specification

for  
**Blaban**  
Version 1.0  

## Prepared by  

- Mohamed moataz 202201469  
- Mohamed ramadan 202201773  
- Amr Mamdouh 202201393  
- Youssef helal 202201863  

## Zewail City

---

# Table of Contents

1. [Introduction](#1-introduction)  
   1.1 [Purpose](#11-purpose)  
   1.2 [Intended Audience](#12-intended-audience)  
   1.3 [Project Scope](#13-project-scope)  

2. [Overall Description](#2-overall-description)  
   2.1 [Product Features](#21-product-features-functional-and-non-functional-requirements)  
   2.2 [User Classes and Characteristics](#22-user-classes-and-characteristics)  
   2.3 [Design and Implementation Constraints](#23-design-and-implementation-constraints)  

3. [System Models](#3-system-models)  
   3.1 [Use Case Models](#31-use-case-models)  
   3.2 [User Stories](#32-user-stories)  

4. [System Diagrams](#4-system-diagrams)  
   4.1 [Context Diagram](#41-context-diagram)  
   4.2 [Container Diagram](#42-container-diagram)  
   4.3 [Component Diagram](#43-component-diagram)  

5. [System ADRs](#5-system-adrs)  
6. [Resilience Strategies](#7-resilience-strategies)  

---

## 1. Introduction

### 1.1 Purpose

Blaban is an Egyptian restaurant designed to provide guests with a genuine taste of Egypt through its diverse menu of traditional dishes. Its primary purpose is to offer an enjoyable and immersive dining experience by allowing patrons to explore the rich flavors of Egyptian cuisine, featuring signature dishes like koshari, crepes, om ali, El Qashtouta, juices all crafted from fresh, high-quality ingredients.

### 1.2 Intended Audience

The primary audiences of this document are developers who will be implementing the software and stakeholders who will be reviewing the functionalities of the program they want and decide which features are to be included in the project.

### 1.3 Project Scope

Our project is a replication of Blaban restaurant, but on a smaller scale.

---

## 2. Overall Description

### 2.1 Product Features (Functional and Non-functional Requirements)

Our platform should have the following features:

#### Functional Requirements:

1. **User Account Management**: the user should be able to sign up and login to the system.
2. **Managing Items**: The user should be able to add/remove items to cart from the menu.
3. **Checkout**: The user should be able to check out the order if there are items in the cart.
4. **Reviews**: The user should be able to leave reviews on the site after making an order.
5. **Total reviews**: the system should count the total number of reviews and the average based on multiple users.
6. **Filter**: The user should be able to filter items based on price and category.
7. **Search:** users should be able to search for a specific item by name.
8. **Discounts:** The manager should be able to apply discounts on specified items.
9. **Creating New items:** The manager should be able to create new items for users in the menu.
10. **Order History:** The user should be able to view their past orders, including details such as items ordered, order date, and total amount spent.

#### Non-functional Requirements:

1. **Response time:** the system shall process requests and load pages within a max of 3 seconds.
2. **Usability:** user should be able to do a task within 2-3 clicks from the homepage.
3. **Performance:** The system shall have a below 3 seconds response time.
4. **Maintainability:** The codebase should be modular, applying single responsibility to allow for extendable modifications and updates regularly.
5. **Security:** The system should authenticate and authorize users using OAuth2.

---

### 2.2 User Classes and Characteristics

- **Users:** Users of the restaurant application can explore a diverse menu featuring a variety of dishes and beverages. They can place orders, discover new menu items. The platform enhances the dining experience by allowing users to search for their favorite item and filter based on different categories and financial conditions.

- **Managers:** Managers of the restaurant application can oversee the menu by adding new dishes and beverages, updating existing items, and managing pricing. They have the ability to analyze customer preferences to optimize the menu offerings, and can also engage with users by adding discounts to attract more customers.

### 2.3 Design and Implementation Constraints

This project will be developed as an online web application designed to be simple and user-friendly. The primary focus will be on providing a seamless and straightforward experience while effectively implementing essential functionalities such as ordering food.

---

## 3. System Models

### 3.1 Use Case Models

#### Use Case ID: UC-01

**Use case name:** User registration (signup)

**Actors:**
- Main actors: user  
- Secondary actors: system  

**PreConditions:**
- User has an internet connection and is on the registration page.

**Main Flow:**
1. User enters required fields (name, password).  
2. User clicks on the sign up button.  
3. System checks if the user entered invalid name or password (e.g., entering only numbers in the name section).  
4. User is routed to the login page.  

**Alternative Flow:**
- 3a. If User password or name is incorrect, it displays an error message.
- 3b. Internet connection error occurs when the user or system cannot communicate correctly.

**PostConditions:**
- An account is created successfully and the user can now login.

---

#### Use Case ID: UC-02

**Use Case Name:** Place Food Order

**Actors:**
- **Main Actor:** User  
- **Secondary Actor:** System  

**Preconditions:**
User has an internet connection and is logged into their account. The user has items in their cart.

**Main Flow:**
1. User navigates to the cart page.  
2. User reviews the items in the cart and confirms the order details (quantities, total price).  
3. User clicks on the "Checkout" button.  
4. System processes the order and confirms payment details.
5. System displays an order confirmation message with estimated delivery/pickup time.

**Alternative Flow:**
- 4a. If the payment fails, the system displays an error message and prompts the user to re-enter payment information.
- 4b. If there are no items in the cart, the system displays a message prompting the user to add items before checking out.
- 4c. If an internet connection error occurs, the system displays an error message and prompts the user to check their connection.

**Postconditions:**
The order is successfully placed, and the user receives an order confirmation with details.

---

#### Use Case ID: UC-03

**Use Case Name:** Add Discount to Menu Item

**Actors:**
- **Main Actor:** Manager
- **Secondary Actor:** System

**Preconditions:**
The manager is logged into their account and has access to the menu management interface.

**Main Flow:**
1. Manager navigates to the menu management section of the application.
2. Manager selects the item to which they want to apply a discount.
3. Manager enters the discount percentage or amount in the designated field.
4. Manager clicks on the "Apply Discount" button.
5. System validates the discount details and updates the item in the menu.
6. System displays a confirmation message indicating that the discount has been successfully applied.

**Alternative Flow:**
- 5a. If the discount exceeds the item's price, the system displays an error message indicating that the discount cannot be greater than the item price.
- 5b. If an internet connection error occurs, the system displays an error message and prompts the manager to check their connection.

**Postconditions:**
The discount is successfully applied to the selected menu item, and the item is updated in the system for users to see.

---

### 3.2 User Stories

#### 1. Customer User Stories:

1.1. As a user, I want to create an account so that I can order food.  
1.2. As a user, I want to be able to login to an existing account so that I can access my details.  
1.3. As a user, I want to be able to access the menu so that I can add or drop items.  
1.4. As a user, I want to be able to search and filter for a specific item so that I can quickly find what I am looking for.  
1.5. As a user, I want to add items to cart so that I can buy them.  
1.6. As a user, I want to be able to leave reviews on items so that I can share my experience with others.  
1.7. As a user, I want to proceed to checkout and make finalisations so that I can add or drop more items before ordering.  
1.8. As a user, I want to be able to use discount codes on an item so that I can buy them for a cheaper price.  

#### 2. Manager User Stories:

2.1. As a Manager, I want to create new items so that I can add them to menu.  
2.2. As a Manager, I want to be able to update existing items so that I can update price or info when needed.  
2.3. As a Manager, I want to be able to apply discounts on a specific item so that I can attract more customers and increase sales.  

---






## 5. System ADRs

### ADR 001: Microservices Architecture

**Context**: we need a robust, scalable and maintainable system which has its features integrated separately.

**Decision**: adopt a microservices architecture using **Golang**.

**Consequences**:

#### Pros:
- Improved scalability
- Independent service deployment

#### Cons:
- Distributed system challenges
- High cost
- Increased complexity

---

### ADR 002: OAuth2 Authentication

**Context**: we need a secured and simple way to authorize users and authenticate them.

**Decision**: users authentication and authorization will be implemented using OAuth2 (google authentication).

**Consequences**:

#### Pros:
- Reducing password handling complexities
- Secured token based authentication

#### Cons:
- Relying on third party providers

---

### ADR 003: API Gateway Implementation

**Context:** need for centralized routing middleware which has high performance and easy routing.

**Decision:** use Fiber go as the main API for the system.

**Consequences:**

#### Pros:
- Extremely low memory allocation
- Near-native performance
- Easy to learn and use
- Strong middleware support

#### Cons:
- Relatively new framework
- Limited enterprise-level features out of the box

---

### ADR 004: Use of MongoDB as the Primary Database

**Context:** the restaurant system requires a flexible schema-less database to store dynamic user data, reviews and restaurant menus.

**Decision:** We have chosen MongoDB as the primary database for the system.

**Consequences:**

#### Pros:
- NoSQL flexibility which allows for easy storage and retrieval of data
- MongoDB supports horizontal scaling which is beneficial for handling large datasets.
- Replica sets in MongoDB should ensure high availability and fault tolerance.

#### Cons:
- May require additional indexing strategies for optimal query performance

---

### ADR 005: Use of Redis for Caching

**Context:** To improve system responsiveness and reduce load on the Database, We need a caching mechanism for frequently accessed data.

**Decision:** We have chosen Redis as the caching layer.

**Consequences:**

#### Pros:
- Redis operating in memory should allow fast read/write operations.
- Redis session management will be useful for user authentication and storing session data.
- Redis supports clustering, ensuring high availability and distributed caching and fault tolerance.

#### Cons:
- Requires cache invalidation policies to prevent serving stale data.

---

### ADR 006: Use of Grafana and Prometheus for Logging and Monitoring

**Context:** To ensure observability, performance monitoring and system health tracking, We need a robust logging and monitoring system to track all the microservices we have.

**Decision:** We have chosen Grafana and Prometheus for logging and monitoring due to their ability to handle real-time metrics efficiently.

**Consequences:**

#### Pros:
- Grafana provides a powerful visualization dashboard that allows real-time monitoring of system health and performance metrics.
- Prometheus acts as the metrics collection system storing time-series data and integrating it seamlessly with Grafana.
- Prometheus exporters natively support Go applications making it easy to gather application-level metrics.
- Grafana can be used to notify the operations team in case of system anomalies.
- Prometheus can handle large-scale data collection improving scalability.

#### Cons:
- Potential storage concerns for high-volume logging.

---

### ADR 007: Order Service

**Context:** The Blaban Restaurant system requires a dedicated service for managing order that can be scaled, tested and has high performance.

**Decision:** Implement an Order Microservice using Golang.

**Consequences:**

#### Pros:
- Independent scalability
- Clear separation of concerns
- Easy to maintain and update

#### Cons:
- Increased system complexity
- Requires complex monitoring
- Hard to debug errors and failures

---

### ADR 008: Review Service

**Context:** The Blaban Restaurant system requires a feature that allows users to leave reviews for menu items and if it fails it won't affect menu service.

**Decision:** implement a separated review service.

**Consequences:**

#### Pros:
- Increased resilience  
- Maintainable  

#### Cons:
- Additional service adds deployment and management complexity.  
- Requires API communication between the menu service and the review service.  

---

### ADR 009: Payment Service

**Context:** The Blaban Restaurant requires an external payment microservice to integrate with our application in order for customers to be able to pay for the orders and be separate from handling orders.

**Decision:** Integrate an independent payment microservice with our system.

**Consequences:**

#### Pros:
- Increased resilience  
- Independent Scalability

#### Cons:
- Reliance on a third party system.

---

## 7. Resilience Strategies

- To ensure high availability and improve read performance, we will use database replicas for the review service. This ensures that even if the primary database fails, read requests can still be served by the replicas, also there is a feature in MongoDB to replicate databases.

- To automate service recovery and minimize downtime, we will use Docker's restart policies. If the review service crashes, Docker will automatically restart it, ensuring minimal disruption. This approach simplifies deployment and helps maintain system stability without requiring manual intervention.

- To ensure continued operation if one API gateway instance fails we will implement a failover mechanism by deploying multiple instances of the API Gateway behind a load balancer.

- To prevent cascading failures when a service is experiencing high latency or is down, we will implement a circuit breaker to detect failures and temporarily block requests to failing services.

- To prevent overloading services due to excessive requests and to prevent abuse, we will use API gateway throttling through Nginx rate limiting on requests.

- To ensure requests are evenly distributed preventing a single instance from becoming a bottleneck, we will deploy multiple backup instances of microservices and use a load balancer (e.g. Nginx or HAProxy).

- To reduce load on the database, we will serve frequently accessed data from an in-memory cache by implementing Redis.