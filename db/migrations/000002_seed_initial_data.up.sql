-- Seed 6 conversations
INSERT INTO conversations (chat_id, user_id, state, last_interaction_at) VALUES
(2001, 3001, 'Idle', NOW() - interval '10 minutes'),
(2002, 3002, 'Idle', NOW() - interval '20 minutes'),
(2003, 3003, 'Idle', NOW() - interval '30 minutes'),
(2004, 3004, 'AwaitingReview', NOW() - interval '5 minutes'), -- Start in different state
(2005, 3005, 'Idle', NOW() - interval '1 hour'),
(2006, 3006, 'Idle', NOW() - interval '2 hours')
ON CONFLICT (chat_id) DO NOTHING;

-- Seed 6 messages for chat_id 2001
INSERT INTO message_history (chat_id, is_user_message, text, "timestamp") VALUES
(2001, TRUE,  'Hi, can you tell me your opening hours?', NOW() - interval '9 minutes'),
(2001, FALSE, 'We are open from 9 AM to 5 PM, Monday to Friday.', NOW() - interval '8 minutes'),
(2001, TRUE,  'Okay, great. And are you open on Saturdays?', NOW() - interval '7 minutes'),
(2001, FALSE, 'No, we are closed on weekends.', NOW() - interval '6 minutes'),
(2001, TRUE,  'Alright, thank you very much!', NOW() - interval '5 minutes'),
(2001, FALSE, 'You''re welcome! Is there anything else I can help with?', NOW() - interval '4 minutes');

-- Seed 6 messages for chat_id 2002
INSERT INTO message_history (chat_id, is_user_message, text, "timestamp") VALUES
(2002, TRUE,  'I need to return an item.', NOW() - interval '19 minutes'),
(2002, FALSE, 'I can help with that. What is the item and order number?', NOW() - interval '18 minutes'),
(2002, TRUE,  'It''s the blue widget, order #5678.', NOW() - interval '17 minutes'),
(2002, FALSE, 'Thank you. Please send it back to our main address with the return slip.', NOW() - interval '16 minutes'),
(2002, TRUE,  'Will do, thanks.', NOW() - interval '15 minutes'),
(2002, FALSE, 'No problem! Let us know if you have more questions.', NOW() - interval '14 minutes');

-- Seed 6 messages for chat_id 2003
INSERT INTO message_history (chat_id, is_user_message, text, "timestamp") VALUES
(2003, TRUE,  'Where are you located?', NOW() - interval '29 minutes'),
(2003, FALSE, 'Our main office is at 123 Main Street.', NOW() - interval '28 minutes'),
(2003, TRUE,  'Is there parking available nearby?', NOW() - interval '27 minutes'),
(2003, FALSE, 'Yes, there is a public parking garage one block west.', NOW() - interval '26 minutes'),
(2003, TRUE,  'Perfect, thanks for the info!', NOW() - interval '25 minutes'),
(2003, FALSE, 'Happy to assist!', NOW() - interval '24 minutes');

-- Seed 6 messages for chat_id 2004 (starts AwaitingReview, maybe they just said thanks)
INSERT INTO message_history (chat_id, is_user_message, text, "timestamp") VALUES
(2004, TRUE,  'What''s the status of my repair?', NOW() - interval '10 minutes'),
(2004, FALSE, 'Let me check... It looks like repair #9012 is complete and ready for pickup!', NOW() - interval '9 minutes'),
(2004, TRUE,  'Fantastic! That was quick.', NOW() - interval '8 minutes'),
(2004, FALSE, 'We try our best! You can pick it up anytime during business hours.', NOW() - interval '7 minutes'),
(2004, TRUE,  'Awesome, thank you so much!', NOW() - interval '6 minutes'),
(2004, FALSE, 'You''re most welcome! We''d love to hear about your experience if you have a moment.', NOW() - interval '5 minutes'); -- Bot asks for review

-- Seed 6 messages for chat_id 2005
INSERT INTO message_history (chat_id, is_user_message, text, "timestamp") VALUES
(2005, TRUE,  'Is the X1 model in stock?', NOW() - interval '59 minutes'),
(2005, FALSE, 'Let me check inventory... Yes, we have 3 units of the X1 in stock right now.', NOW() - interval '58 minutes'),
(2005, TRUE,  'Can I reserve one?', NOW() - interval '57 minutes'),
(2005, FALSE, 'Unfortunately, we don''t offer reservations, it''s first-come, first-served.', NOW() - interval '56 minutes'),
(2005, TRUE,  'Okay, I understand. I''ll try to come by soon.', NOW() - interval '55 minutes'),
(2005, FALSE, 'Sounds good! Hope to see you.', NOW() - interval '54 minutes');

-- Seed 6 messages for chat_id 2006
INSERT INTO message_history (chat_id, is_user_message, text, "timestamp") VALUES
(2006, TRUE,  'Do you offer gift wrapping?', NOW() - interval '1 hour 59 minutes'),
(2006, FALSE, 'Yes, we do offer complimentary gift wrapping for most items!', NOW() - interval '1 hour 58 minutes'),
(2006, TRUE,  'Great! Can I request it when I order online?', NOW() - interval '1 hour 57 minutes'),
(2006, FALSE, 'Absolutely. There should be a checkbox or option during checkout.', NOW() - interval '1 hour 56 minutes'),
(2006, TRUE,  'Excellent, thank you!', NOW() - interval '1 hour 55 minutes'),
(2006, FALSE, 'You''re welcome! Happy gifting!', NOW() - interval '1 hour 54 minutes');