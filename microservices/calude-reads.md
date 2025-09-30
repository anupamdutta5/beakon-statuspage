# Guidelines for AI Code Editor

## 1. Microservice Boundaries
- Each feature belongs to a **specific microservice**.  
- Always implement features in the **correct microservice** and then establish the necessary connections between services.  
- Avoid mixing responsibilities across microservices.  

## 2. Avoid Duplication
- Before writing new code, check whether the feature or function is already implemented in **any service**.  
- Reuse existing functionality where possible instead of duplicating logic.  

## 3. Understand the Architecture
- Explore the `microservices/` directory.  
- Review **README files** of each microservice to understand its purpose, APIs, and dependencies.  
- Update documentation if you add or change functionality.  

## 4. Code Quality & Practices
- Avoid “quick fixes” — always use **clean, maintainable solutions**.  
- Follow established **best practices** for coding, error handling, and service communication.  
- Write tests where appropriate to ensure reliability.  

## 5. Robustness & Resilience
- Design with **fault tolerance** in mind (graceful failure, retries, timeouts).  
- Keep services **loosely coupled** but **well-integrated**.  
- Ensure that services can recover quickly from crashes or errors.  

## 6. File & Code Organization
- Split large files into **smaller, modular units** for better readability and maintainability.  
- Keep functions/classes focused on **single responsibilities**.  

## 7. Documentation & Communication
- Always update documentation after implementing or modifying features.  
- Provide clear commit messages and code comments to help future contributors.  

## 8. Performance & Scalability
- Consider the impact of your changes on **database load**, **API latency**, and **scalability**.  
- Use caching, batching, or async operations where needed.  

## 9. Security & Compliance
- Validate all inputs and handle sensitive data securely.  
- Follow least-privilege principles for service-to-service communication.  
- Avoid exposing internal details through APIs.  

## 10. Consistency Across Services
- Use consistent **naming conventions**, **logging standards**, and **error-handling patterns**.  
- Keep configs, secrets, and environment variables managed properly (e.g., via ConfigMaps/Secrets).  
