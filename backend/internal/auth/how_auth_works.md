1. **Frontend → PocketBase**

   * Login via frontend directly with pb.
   * PocketBase returns a JWT (`token`) and the user record.

2. **Store the JWT client-side**

3. **All subsequent requests to your Go game API**

   * Include the header

     ```txt
     Authorization: Bearer <token>
     ```

   * The middleware forwards that token to PocketBase's `auth-refresh` endpoint and, on success, injects the player's ID into the request context for game logic.

auth is frontend only, talks to PocketBase for login/refresh; every game call carries the same JWT to the backend to keep track of the player.
