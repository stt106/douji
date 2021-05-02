# Douji
"Douji" is a simple poker game which I find interesting; as I recently started learning Go I decided to implement it with Go as a learning project.

## Rules
1. There is no dealer, it requires at least two players to play against all other players. Before each game starts, it costs each player a base point which is usually customised to be the minimum calling point in the game. So each game always starts with some points in the "pot"!

2. Numbered card's score is its rank, e.g. a card with rank 6 has a score of 6 regardless of its suit.
    J scores 11 point, Q scores 12, K scores 13 and A scores 15; again suit doesn't matter.
    black Joker scores 17 points while red Joker scores 19 points.
    There is also an extra card called "special card", it doesn't have any rank or suit, its score is 21 points. This card normally comes with a brand new deck used as a branding card from the card producer but now it has been put into a good use:)

3. There are mainly two versions of the game\n:
    V1. Each player starts with one hidden card (visible to himself/herself only until the final round) and one public card (visible to everyone).
    V2. Each player starts with two hidden cards, no public card.
    After it starts, in V1, there are 4 rounds and in V2 there are 5 rounds. In both cases, there is no card dealing (since there are hidden cards) in the final round where players compares final score taking into account of the hidden cards. Before the final round, for any player stays in the game, a new public card is dealt. 
    So in V1, a player at most gets a hand of 5 cards (1 hidden and 4 public) and in V2 a player at most gets a hand of 6 cards (2 hidden and 4 public).

4. Scores:
    Face score is each player's *last* public card score.
    A player's public score is the sum of all public cards' score. 
    A player's final score is the sum of all cards score, including both hidden and public cards.

5. At the final round, a player with the largest final score wins the game. When there is a tie in the final score; it's called "bombed pot" meaning there is no winner in the current game; it continues to the next game where on base point is needed from each player, the bombed pot from the last game is the starting pot of the new game. This continues until there is a winner in the current game and the next game starts as normal. It's rare but possible to have consecutive bobmed pots; this is another reason why I find this game interesting. You may lose before the final round but you never know what happens until the very end!

6. After the game starts, each round a calling player calls either 0, 1, 2, 3, 4, or 5 points, where 0 indicates quitting the game. In the final round, there is an extra calling option of 10 points. Logically this is because only in the final round, a player can see other players' hidden card provided they remain in the game at the final round. Hence even if a player with a small public score can call large points in the final round if he/she knows his/her final score will be larger than other players'. Another case where the final round calling is larger is that a player wants to trick others by pretending having a large hidden card, hoping others won't want to call the game to see his/her hidden card. Naturally this is the interesting part of the game and it requires both courage and skills.

7. ### Calling rule of each round. 
    * With V1, it's always the player with the largest *face score* calls the round. If there is a tie on the face score, starting from the last calling player, counter-clockwise, the first one with the largest face score calls this round. 
    * With V2, since there is no public card initially, the previous game winner calls the first round. If it's the first game i.e. there is on previous winner, the first player joins the game calls the first round. From 2nd round onwards where there are public cards, it uses the largest face score calling rule used in V1. 

8. In each round, once a calling player calls a certain point by increasing the pot, other players either choose In or Out. If choosing In, he/she has to add the calling point to the pot. Otherwise, the player loses everything he/she adds to the pot (unless it's bombed pot, the lost player gets another chance!). Naturally, when someone calls, no other player chooses In, then the game is over and the calling player wins the pot. So final score is only relevant in the final round. 


9. ### Extra point rules
 If the game is just like this, it would be less interesting. There are rules which lead to extra points for a hand.
  1. Wild card: Spade-2. Wild card is special because it can magically become another card in need hence increase the score a lot. The choice of spade-2 as the wild card is "clever" as spade-2 normally only has 2 points so players don't like it but because of its speciality, players also want it!
  2. In one deck of cards, a hand of five a kind rules rules everything by having a const final score of **300** points, even with extra point rule no other hand can possible beat 300 points! To get five a kind in one deck, you must have four a kind plus the wild card s.a. 6666+spade 2 where wild card becomes the 5th 6!
  3. A hand of four a kind gets extra **60** points, even with one deck, it's possible (though extremely rare) to get multiple hands of four a kind.
   When there is **only one** hand of four a kind in the final round, it wins the game and final score doesn't matter. 
   When there is more than one hand of four a kind in the final round, it still compares the final score of each hand applying standard scoring rule to determine a winner, s.a. 555+wild card doesn't auto win when someone else has a hand of 44446
  3. Two Jokers gets at least extra **30** points.
    Again, wild card can play a role here; so one joker + wild card is equivalent to two jokers, depending on which pure joker in hand, wild card becomes the other joker hence in addition to the 30 extra points, wild card itslef also gets extra 15 or 17 points. 
    In the case of having two jokers and wild card, the wild card becomes the red Joker (as this results the max point increase.)
    Remember there is a special card having 21 points? Special card with wild card does NOT get any extra points.
   4. Each three a kind gets extra 30 points. 
    Note that in V1 where there are at most 5 cards in a hand, it's only possible to have one three a kind but in V2, it's possible to have two three a kind (s.a. 666777 or 44455wildcard). 
    When it's possible to use the wild card to get either double jokers or three a kind, it's always preferable to get double jokers since this results the max point increase. So a hand of one joker, wild card, 3, 3, 5 shall treat the wild card a joker rather a card of rank 3.
    Also, when there are two pairs with a wild card, pick the pair with higher rank to get a three a kind, again this results the max point increase. E.g. 4455wild card shall treat the wild card as 5 rather than 4.
    
