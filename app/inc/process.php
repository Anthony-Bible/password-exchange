<?php
//namespace SendGrid;
require (__DIR__ . '/../vendor/autoload.php');

$dotenv = Dotenv\Dotenv::createImmutable(__DIR__ . '/../');
$dotenv->load();

use \Aws\Ses\SesClient;
use \Aws\Credentials\CredentialProvider;
use \Aws\Exception\AwsException;
print_r($_POST)

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
		$phone=$_POST["phone"];
		$message = $_POST["message"];
		$subject = "Thanks for contacting me";
		$plaintext_body = 'This email was sent with Amazon SES using the AWS SDK for PHP.' ;
		$html_body =  '';
		$char_set = 'UTF-8';
	

	try 
		{
			/* We've set all the 	, it's now time to send it. To do this we just check the captcha response. If they failed we won't send the mail. This has dramatically reduced the spam to almost zero */
			
			$secret= getenv('GOOGLECAPTCHASECRET');
			$captchaResponse=$_POST["captcha"];
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
