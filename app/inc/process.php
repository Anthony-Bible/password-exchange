<?php
//namespace SendGrid;
require (__DIR__ . '/../vendor/autoload.php');
include_once(dirname(__FILE__) . "/database.php"); 
include_once(dirname(__FILE__) . "/functions.php");

$dotenv = Dotenv\Dotenv::createImmutable(__DIR__ . '/../');
$dotenv->load();

use \Aws\Ses\SesClient;
use \Aws\Credentials\CredentialProvider;
use \Aws\Exception\AwsException;
#print_r($_POST);
### STEPS ###
# 1. Generate Random string
# 2. Encrypt message with random string
# 3. Store Message and recpients info in db
# 4. Send email with link and random string
# 5. User gets to page
# 6. User enters random string, first and last name
# 7. We verify it's them
# 8. Take them to the screen 
# 9. Set thee data for deletion
# 10. Clean data every 7 days
// print_r($_POST);
$firstname=$_POST['other_firstname'];
$lastname=$_POST['other_lastname'];
$token = RandomToken(64);
list($iv, $encryptedstring, $key) = EncryptMessage($_POST['message'],$token);
echo $encryptedstring;
print_r($_ENV);
storeMessageinDatabase($encryptedstring, $firstname, $lastname);
#Array ( [name] => Anthony Bible [email] => pine64@anthony.bible [other_firstname] => Anthony [other_lastname] => Bible [other_email] => test2@anthony.bible [message] => yeat a nother test 1268 [g-recaptcha-response] => 03AGdBq27UFdxuDUc_IPa-9r4AYLLiDq-zYaCC2oTyKRRGoDW7D2-_bfPX9Vg14kI_-Z7-cinNq3vzHCTeNl4hofRopK9-SYaFIUNyzOxqj9z9hFWD9rhBF3mZ3YFujFgiEiUDztUifvkKQ-C0rXyB-fM1YB8Nq3apv0KpcIDZI4xc0A9QHu5SfuscGuJQ-X_vj9yIVwyDbnT6cY_untqkLcHXTV5q_Jz4vWuNglHDpr7d7yuMptXao9pdU71sTpAZCuan8nqMz2hNn34Me-2loPxguilSBbqAIhsRxKoW1vHTi8jOvhJvT8b2hErMKKh9hTMh6118riofIX-7C5zZng2BskPIIo6yU-hK9pnFADJA_4AIPEuUvupwQaVDrnwzzjqUZO1hqOwO1neg7A3XsILeWaEfstvFow2QBy-AY8oVpU6QSv79MR7I7VcJkmoXEp5CzEYBt1Mvt4AXuYjUOn28z-gjtS7rHw )
function sendEmail(){
	### TODO ### 
    # Create function with all variables in an array #
	
	$provider = CredentialProvider::env();
echo '<?xml version="1.0" encoding="UTF-8" ?>'; 
$SesClient = new SesClient([
    'version' => 'latest',
	'region'  => 'us-west-2',
	'credentials' => $provider
]);
$sender_email = getenv('SENDEREMAIL');

	
		$recieverEmail=[$_POST["email"], "anthony@anthonybible.com"];
		$receiverid = $_POST["name"];
		$subject = "Thanks for contacting me";
		$plaintext_body = 'This email was sent with Amazon SES using the AWS SDK for PHP.' ;
		$html_body =  '';
		$char_set = 'UTF-8';
	

	try 
		{
			/* We've set all the 	, it's now time to send it. To do this we just check the captcha response. If they failed we won't send the mail. This has dramatically reduced the spam to almost zero */
			
			$secret= getenv('GOOGLECAPTCHASECRET');
			$captchaResponse=$_POST["g-recaptcha-response"];
			echo "<response>";
			echo "<message>";

			$verifyUrl="https://www.google.com/recaptcha/api/siteverify?secret=$secret&response=$captchaResponse";
			$verify=file_get_contents($verifyUrl);
				$captcha_success=json_decode($verify);
				if ($captcha_success->success==false) {			
					echo "Looks like the robot overlords deterimined you were a bot, please try the Recaptcha again";
				}
				
				if ($captcha_success->success==true) {
				//This user is verified by recaptcha
				$result = $SesClient->sendEmail([
					'Destination' => [
						'ToAddresses' => $recieverEmail,
					],
					'ReplyToAddresses' => [$sender_email],
					'Source' => $sender_email,
					'Message' => [
					  'Body' => [
						  'Html' => [
							  'Charset' => $char_set,
							  'Data' => $html_body,
						  ],
						  'Text' => [
							  'Charset' => $char_set,
							  'Data' => $plaintext_body,
						  ],
					  ],
					  'Subject' => [
						  'Charset' => $char_set,
						  'Data' => $subject,
					  ],
					],
					
				]);
				$messageId = $result['MessageId'];
				echo "<h3>You Successfully sent the Email if you don't recieve an email please check your spam folder</h3>";



				}

			
			echo "</message>";
			echo "</response>";
			

		}
	 catch (Exception $e) {
	
			echo "<response>";
			echo "<message>";
			// output error message if fails
			echo $e->getMessage();
			echo "\n";
			echo "</message>";
			echo "</response>";
			
	}
}
sendEmail();










?>
