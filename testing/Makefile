MSG = 'Hello, World!'
MSG2 = 'attacking tonight'
KEY = PASSWORD
KEY2 = oculorhinolaryngology
IP = 127.0.0.1
PORT = 9876
MIN_DELAY = -m 2
MAX_DELAY = -M 4

server: clear
	./server $(IP) $(PORT)

t1: clear
	@echo "Test 1: Server w/ no args"
	./server

t2: clear
	@echo "Test 2: Server -h"
	./server -h

t3: clear
	@echo "Test 3: Server w/ standard args"
	./server $(MIN_DELAY) $(MAX_DELAY) $(IP) $(PORT)

t4: clear
	@echo "Test 4: Server not enough args"
	./server $(IP)

t5: clear
	@echo "Test 5: Server args out of order"
	./server $(PORT) $(IP)

t6: clear
	@echo "Test 6: Server w/ invalid ip"
	@echo "Case 1"
	./server 127.987 $(PORT)
	@echo "Case 2"
	./server 127.0.0.1.1.1 $(PORT)
	@echo "Case 3"
	./server 300.300.300.300 $(PORT)

t7: clear
	@echo "Test 7: Server w/ invalid port"
	@echo "Case 1: Negative"
	./server $(IP) -1
	@echo "Case 2: Over max range"
	./server $(IP) 99999

t8: clear
	@echo "Test 8: Server w/ invalid min/max delay"
	@echo "Case 1: min floating point number"
	./server -m 2.5 -M 4 $(IP) $(PORT)
	@echo "Case 2: max floating point number"
	./server -m 2 -M 4.5 $(IP) $(PORT)
	
t9: clear
	@echo "Test 9: Server w/ extra args"
	./server $(IP) $(PORT) extra args

t10: clear
	@echo "Test 10: Server graceful shutdown"
	./server

t11: clear
	@echo "Test 11: Server handle multiple clients"
	./server

t11client: clear
	@echo "Test 11: Server handle multiple clients"
	@echo "Set 1"
	./client $(MSG) $(KEY) $(IP) $(PORT)
	@echo "Set 2"
	./client $(MSG2) $(KEY2) $(IP) $(PORT)

t12: clear
	@echo "Test 12: Client w/ no args"
	./client

t13: clear
	@echo "Test 13: Client -h"
	./client -h

t14: clear
	@echo "Test 14: Client w/ standard args MSG KEY IP PORT"
	@echo "Set 1"
	./client $(MSG) $(KEY) $(IP) $(PORT)
	@echo "Set 2"
	./client $(MSG2) $(KEY2) $(IP) $(PORT)

t15: clear
	@echo "Test 15: Client w/ not enough args"
	@echo "Case 1: Only MSG"
	./client $(MSG)
	@echo ""
	@echo "Case 2: MSG and KEY (Missing IP and PORT)"
	./client $(MSG) $(KEY)
	@echo ""
	@echo "Case 3: MSG, KEY, and IP (missing PORT)"
	./client $(MSG) $(KEY) $(IP)
	
t16: clear
	@echo "Test 16: Client args out of order"
	@echo "Case: IP PORT MSG KEY"
	./client $(IP) $(PORT) $(MSG) $(KEY)

t17: clear
	@echo "Test 17: Client w/ special characters in MSG"
	@echo "Case 1"
	./client 'Hello, World! !@#\$%' $(KEY) $(IP) $(PORT)
	@echo ""
	@echo "Case 2: Only special characters"
	./client '!@#\$%^&*()' $(KEY) $(IP) $(PORT)

t18: clear
	@echo "Test 18: Client w/ numbers in MSG"
	@echo "Case 1"
	./client 'Hello, World! 12345' $(KEY) $(IP) $(PORT)
	@echo ""
	@echo "Case 2: Only numbers"
	./client '123456789' $(KEY) $(IP) $(PORT)

t19: clear
	@echo "Test 19: Client w/ special characters in KEY"
	@echo "Case 1"
	./client $(MSG) 'PASSWORD!@#\$%' $(IP) $(PORT)
	@echo "Case 2: Only special characters"
	./client $(MSG) '!@#\$%^&*()' $(IP) $(PORT)

t20: clear
	@echo "Test 20: Client w/ numbers in KEY"
	@echo "Case 1"
	./client $(MSG) PASSWORD12345 $(IP) $(PORT)
	@echo "Case 2: Only numbers in KEY"
	./client $(MSG) 123456789 $(IP) $(PORT)

t21: clear
	@echo "Test 21: Client w/ invalid IP"
	@echo "Case 1"
	./client $(MSG) $(KEY) 127.0.0.1.1 $(PORT)
	@echo "Case 2"
	./client $(MSG) $(KEY) 127.0 $(PORT)
	@echo "Case 3"
	./client $(MSG) $(KEY) 300.0.0.1 $(PORT)

t22: clear
	@echo "Test 22: Client w/ invalid PORT"
	@echo "Case 1: Negative"
	./client $(MSG) $(KEY) $(IP) -1
	@echo "Case 2: Over max range"
	./client $(MSG) $(KEY) $(IP) 99999

t23: clear
	@echo "Test 23: Client w/ too many arg"
	./client $(MSG) $(KEY) $(IP) $(PORT) extra args

t24: clear
	@echo "Test 24: Client run without Server online"
	./client

clear:
	@clear
