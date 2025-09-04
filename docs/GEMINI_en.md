# Gemini CLI Action Rules

The rules specified in this document outline what the Gemini CLI should and should not do. These rules can be added to, deleted, or changed at any time and serve as an important standard for defining the behavior of the Gemini CLI. The Gemini CLI must always check and follow the rules specified below before taking any action.

## Things to Do

- Always check and act according to the rules specified in the `docs/GEMINI_en.md` document.
- Converse with the user in English. This is to facilitate smooth communication with the user.
- For the command "Reload project context," read `README.md` > `docs/development_plan.md`, then read `**/*.go` files to understand the current state of the project, and ask additional questions if necessary to ensure a clear understanding.
- All feature additions will be done in a TDD fashion. Therefore, always write test cases before adding a feature, confirm that the test fails, and then implement the feature. After implementation, confirm that the test passes.
- When refactoring, also write as many diverse test cases as possible for the logic being refactored, and confirm that all tests pass after the refactoring.
- All comments and documentation within the code should be written in English. This is to improve code readability and facilitate international collaboration.
- If there are nested `for` loops or complex logic, use `logrus` to log intermediate states so the user can visually check the state changes of various variables. This is to help with debugging and understanding.
- In the future, I will request log-based problem analysis when issues occur, so when adding/modifying code, please leave logs in a format that is as easy as possible for you to analyze.
- When adding tests, the slice containing player names should be written as `[]string{"YOU", "CPU1", "CPU2"}`, where the first player's name is always "YOU" and the rest are in the "CPUn" format. This is because the current business logic is based on these hardcoded names.

## Things Not to Do

- For any attempt that fails more than twice, stop on the third try, inform the user in detail about what was attempted and what failed, and then wait for the user's command.
- Never arbitrarily delete or modify any logs I have added, regardless of their level. Logs can be important clues for you and me to analyze problems. Do not remove debug logs even during refactoring.
- When requested to create a title or description for a commit or PR, provide it in a format that I can easily copy and paste without line numbers. Refer to the bad and good examples below.
    - Bad:
      ```
      Commit Title:

      1 feat(poker): Implement PLO hand evaluation logic

      Commit Description:

      1 Refactor the hand evaluation logic to support poker variants with specific hole card usage rules, such as Pot-Limit Omaha (PLO).
      2
      3 The main `EvaluateHand` function is now generalized. It generates all possible 5-card hands based on the game's `UseConstraint` rule ("exact" or
      "any") and evaluates each combination to find the best possible hand.
      4
      5 Key changes:
      6 - Extracted the core ranking logic for a 5-card hand into a new `evaluateSingleHand` helper function.
      7 - For "exact" constraint games like PLO, the system now correctly generates combinations by taking exactly 2 cards from the hole and 3 from the community.
      8 - For "any" constraint games, the logic was also updated to use the more robust combination-based evaluation, ensuring consistency.
      9 - Added a new `TestPLOHandEvaluation` test case to drive the implementation and verify the correctness of the PLO logic.
      ```
    - Good:
      ```
      Commit Title:
      
      feat(poker): Implement PLO hand evaluation logic

      Commit Description:
      
      Refactor the hand evaluation logic to support poker variants with specific hole card usage rules, such as Pot-Limit Omaha (PLO).

      The main `EvaluateHand` function is now generalized. It generates all possible 5-card hands based on the game's `UseConstraint` rule ("exact" or "any") and evaluates each combination to find the best possible hand.

      Key changes:
      - Extracted the core ranking logic for a 5-card hand into a new `evaluateSingleHand` helper function.
      - For "exact" constraint games like PLO, the system now correctly generates combinations by taking exactly 2 cards from the hole and 3 from the community.
      - For "any" constraint games, the logic was also updated to use the more robust combination-based evaluation, ensuring consistency.
      - Added a new `TestPLOHandEvaluation` test case to drive the implementation and verify the correctness of the PLO logic.
      ```
