/**
 * App Navigator
 * Main navigation configuration for the wallet app
 */

import React from 'react';
import {NavigationContainer} from '@react-navigation/native';
import {createStackNavigator} from '@react-navigation/stack';
import {createBottomTabNavigator} from '@react-navigation/bottom-tabs';

// Screens
import WelcomeScreen from '../screens/WelcomeScreen';
import CreateWalletScreen from '../screens/CreateWalletScreen';
import ImportWalletScreen from '../screens/ImportWalletScreen';
import HomeScreen from '../screens/HomeScreen';
import SendScreen from '../screens/SendScreen';
import ReceiveScreen from '../screens/ReceiveScreen';
import HistoryScreen from '../screens/HistoryScreen';
import SettingsScreen from '../screens/SettingsScreen';

const Stack = createStackNavigator();
const Tab = createBottomTabNavigator();

/**
 * Main tab navigator for authenticated users
 */
function MainTabs() {
  return (
    <Tab.Navigator
      screenOptions={{
        headerShown: false,
        tabBarStyle: {
          backgroundColor: '#1a1a1a',
          borderTopColor: '#333',
        },
        tabBarActiveTintColor: '#4A90E2',
        tabBarInactiveTintColor: '#888',
      }}>
      <Tab.Screen name="Home" component={HomeScreen} />
      <Tab.Screen name="History" component={HistoryScreen} />
      <Tab.Screen name="Settings" component={SettingsScreen} />
    </Tab.Navigator>
  );
}

/**
 * Root navigator
 */
function AppNavigator() {
  return (
    <NavigationContainer>
      <Stack.Navigator
        initialRouteName="Welcome"
        screenOptions={{
          headerStyle: {
            backgroundColor: '#1a1a1a',
          },
          headerTintColor: '#fff',
          headerTitleStyle: {
            fontWeight: 'bold',
          },
        }}>
        <Stack.Screen
          name="Welcome"
          component={WelcomeScreen}
          options={{headerShown: false}}
        />
        <Stack.Screen
          name="CreateWallet"
          component={CreateWalletScreen}
          options={{title: 'Create Wallet'}}
        />
        <Stack.Screen
          name="ImportWallet"
          component={ImportWalletScreen}
          options={{title: 'Import Wallet'}}
        />
        <Stack.Screen
          name="Main"
          component={MainTabs}
          options={{headerShown: false}}
        />
        <Stack.Screen
          name="Send"
          component={SendScreen}
          options={{title: 'Send Aura'}}
        />
        <Stack.Screen
          name="Receive"
          component={ReceiveScreen}
          options={{title: 'Receive Aura'}}
        />
      </Stack.Navigator>
    </NavigationContainer>
  );
}

export default AppNavigator;
